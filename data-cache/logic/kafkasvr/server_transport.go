package kafkasvr

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
	"trpc.group/trpc-go/trpc-database/kafka"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/transport"
)

var newConsumerGroup = sarama.NewConsumerGroup

func init() {
	transport.RegisterServerTransport("kafka", DefaultServerTransport)
}

var (
	// DefaultServerTransport ServerTransport 默认实现
	DefaultServerTransport = NewServerTransport()

	serviceListenMap = make(map[string]struct{})
	serviceReadyMap  = make(map[string]chan struct{})
	listenMapLock    sync.Mutex
)

// NewServerTransport 创建 serverTransport
func NewServerTransport(opt ...transport.ServerTransportOption) transport.ServerTransport {
	// option 默认值
	kafkaOpts := &transport.ServerTransportOptions{}
	for _, o := range opt {
		o(kafkaOpts)
	}
	return &ServerTransport{opts: kafkaOpts}
}

// WaitReady 等待kafka准备完毕
func WaitReady(serverName string) error {
	var readyChain chan struct{}
	for i := 0; i < 10 && readyChain == nil; i++ {
		listenMapLock.Lock()
		readyChain = serviceReadyMap[serverName]
		listenMapLock.Unlock()
		time.Sleep(time.Millisecond * 100)
	}
	if readyChain == nil {
		return fmt.Errorf("kafka serverName: %v not exist, please check", serverName)
	}
	select {
	case <-readyChain:
		return nil
	}
}

// ServerTransport kafka consumer transport
type ServerTransport struct {
	opts *transport.ServerTransportOptions
}

// ListenAndServe 启动监听，如果监听失败则返回错误
func (s *ServerTransport) ListenAndServe(ctx context.Context, opts ...transport.ListenServeOption) (err error) {
	listenMapLock.Lock()
	defer listenMapLock.Unlock()
	lsOpts := &transport.ListenServeOptions{}
	for _, opt := range opts {
		opt(lsOpts)
	}
	// 判断是否启动过，启动过则不启动
	if _, ok := serviceListenMap[lsOpts.ServiceName]; ok {
		log.Infof("kafka service: %s, %s already start", lsOpts.ServiceName, lsOpts.Address)
		return nil
	}
	log.Infof("kafka user define plugin start service ")

	// 解析出resetSec参数
	resetSec, err := parseResetCfg(lsOpts)
	if err != nil {
		return err
	}

	// 解析 address 中的参数
	kafkaUserConfig, err := kafka.ParseAddress(lsOpts.Address)
	if err != nil {
		return err
	}

	// 获取 sarama 的配置信息
	config := getServerConfig(kafkaUserConfig)

	client, err := sarama.NewClient(kafkaUserConfig.Brokers, config)
	if err != nil {
		return err
	}

	// 创建 custom group  连接 broker，失败会返回错误
	consumerGroup, err := newConsumerGroup(kafkaUserConfig.Brokers, kafkaUserConfig.Group, config)
	if err != nil {
		return err
	}

	handler := newConsumerGroupHandler(ctx, kafkaUserConfig, client, lsOpts, resetSec)

	if len(kafkaUserConfig.Topics) == 0 {
		return errors.New("no topics provided")
	}

	go func() {
		// 服务结束后，关闭 group 和 client
		defer func() {
			if err := consumerGroup.Close(); err != nil {
				log.Errorf("kafka consumerGroup close return err: %s", err)
			}

			log.Debugf("consumerGroup is closed")
		}()

		for {
			// 只有 handler cleanup 时会报错	详见 sarama.consumerGroup.Consume
			if err := consumerGroup.Consume(ctx, kafkaUserConfig.Topics, handler); err != nil {
				log.ErrorContextf(ctx, "kafka server transport: Consume get error:%v", err)
			}

			select {
			case <-ctx.Done(): // 服务已结束，退出服务
				log.ErrorContextf(ctx, "kafka server transport: context done:%v, close", ctx.Err())
				return
			default:
			}
		}
	}()
	serviceListenMap[lsOpts.ServiceName] = struct{}{}
	return nil
}

func newConsumerGroupHandler(ctx context.Context, config *kafka.UserConfig, client sarama.Client, opts *transport.ListenServeOptions,
	resetSec int) sarama.ConsumerGroupHandler {

	limiter := newLimiter(config.RateLimitConfig)
	meta := newMeta(config)

	readyChain := make(chan struct{})
	serviceReadyMap[opts.ServiceName] = readyChain

	// 创建消费者
	if config.BatchConsumeCount > 0 { // TODO 当 Count == 1 时, 是否还有使用 batchConsumer 的必要？
		return &batchConsumerHandler{
			client:        client,
			resetSec:      time.Second * time.Duration(resetSec),
			readyChain:    readyChain,
			opts:          opts,
			ctx:           ctx, // TODO 此处 ctx 似乎没有必要, 直接使用 session.Context 就可以了
			maxNum:        config.BatchConsumeCount,
			flushInterval: config.BatchFlush,
			retryMax:      config.MaxRetry,
			retryInterval: config.RetryInterval,
			trpcMeta:      config.TrpcMeta,
			limiter:       limiter,
			meta:          meta,
		}
	}
	return &singleConsumerHandler{
		client:        client,
		resetSec:      time.Second * time.Duration(resetSec),
		readyChain:    readyChain,
		opts:          opts,
		ctx:           ctx,
		retryMax:      config.MaxRetry,
		retryInterval: config.RetryInterval,
		trpcMeta:      config.TrpcMeta,
		limiter:       limiter,
		meta:          meta,
	}
}

const (
	// appKey和serverKey历史原因，直接使用的string
	appKey           = "overrideCallerApp"
	serverKey        = "overrideCallerServer"
	consumerGroupKey = codec.ContextKey("kafka.config.consumergroup")
	clientIDKey      = codec.ContextKey("kafka.config.client_id")
)

func newMeta(kafkaUserConfig *kafka.UserConfig) codec.CommonMeta {
	return codec.CommonMeta{
		appKey:    "[kafka]",
		serverKey: kafkaUserConfig.Brokers[0],
		// 因为 trpc 要求插件相互不可见，所以 key 冲突是不可避免的
		// 这里不能设置 meta["attr"] = []attribute.KeyValue{...}, 因为semconv有版本，不同插件设置不同版本无法合并上报
		// 不能设置 meta["config"] = interface{}，因为插件不同版本可能造成interface转换失败，导致整个结构不可用。
		// 所以这里最后还是平铺所有的key, 然后监控平台来理解
		consumerGroupKey: kafkaUserConfig.Group,
		clientIDKey:      kafkaUserConfig.ClientID,
	}
}

// newLimiter get a *rate.Limiter
func newLimiter(conf *kafka.RateLimitConfig) *rate.Limiter {
	if conf == nil {
		return rate.NewLimiter(rate.Inf, 0)
	}
	return rate.NewLimiter(rate.Limit(conf.Rate), conf.Burst)
}

type concurrencyLimitHandler struct {
	ch chan struct{}
	h  transport.Handler
}

func (h *concurrencyLimitHandler) Handle(ctx context.Context, req []byte) (rsp []byte, err error) {
	h.ch <- struct{}{}
	defer func() {
		<-h.ch
	}()
	return h.h.Handle(ctx, req)
}

func newConcurrencyLimitHandler(h transport.Handler, limit int) *concurrencyLimitHandler {
	return &concurrencyLimitHandler{
		ch: make(chan struct{}, limit),
		h:  h,
	}
}

func getServerConfig(uc *kafka.UserConfig) *sarama.Config {
	sc := sarama.NewConfig()
	sc.Version = uc.Version
	if uc.ClientID == kafka.DefaultClientID {
		// 兼容旧逻辑，虫洞消费需要 clientID 和 Group 保持一致
		sc.ClientID = uc.Group
	} else {
		sc.ClientID = uc.ClientID
	}

	sc.Metadata.Full = false                // 禁止拉取所有元数据
	sc.Metadata.Retry.Max = 1               // 元数据更新重次次数
	sc.Metadata.Retry.Backoff = time.Second // 元数据更新等待时间

	sc.Net.MaxOpenRequests = uc.NetMaxOpenRequests
	sc.Net.DialTimeout = uc.NetDailTimeout
	sc.Net.ReadTimeout = uc.NetReadTimeout
	sc.Net.WriteTimeout = uc.NetWriteTimeout

	sc.Consumer.MaxProcessingTime = uc.MaxProcessingTime
	sc.Consumer.Fetch.Default = int32(uc.FetchDefault)
	sc.Consumer.Fetch.Max = int32(uc.FetchMax)
	sc.Consumer.Offsets.Initial = uc.Initial
	sc.Consumer.Offsets.AutoCommit.Interval = 3 * time.Second // 定时多久一次提交消费进度
	sc.Consumer.Group.Rebalance.Strategy = uc.Strategy
	sc.Consumer.Group.Rebalance.Timeout = uc.GroupRebalanceTimeout
	sc.Consumer.Group.Rebalance.Retry.Max = uc.GroupRebalanceRetryMax
	sc.Consumer.Group.Session.Timeout = uc.GroupSessionTimeout
	sc.Consumer.MaxWaitTime = uc.MaxWaitTime
	sc.Consumer.IsolationLevel = uc.IsolationLevel
	return sc
}

// 解析出重置offset的配置
func parseResetCfg(lsOpts *transport.ListenServeOptions) (int, error) {
	address := lsOpts.Address
	addressSplit := strings.Split(address, "?")
	if len(addressSplit) != 2 {
		return 0, fmt.Errorf("less param, bad address: %s", address)
	}
	configParams := addressSplit[1]
	values, err := url.ParseQuery(configParams)
	if err != nil {
		return 0, errors.Wrapf(err, " bad address: %s", address)
	}
	resetSec := 0
	if val := values.Get("resetSec"); val != "" {
		if sec, err := strconv.Atoi(val); err == nil && sec >= 0 {
			resetSec = sec
		} else {
			return 0, fmt.Errorf("bad resetSec value, should >= 0")
		}
	}
	// 这里将参数移除掉,否则会后面kafka的参数解会报错
	lsOpts.Address = strings.ReplaceAll(lsOpts.Address, fmt.Sprintf("%s=%d", "resetSec", resetSec), "")
	return resetSec, nil
}
