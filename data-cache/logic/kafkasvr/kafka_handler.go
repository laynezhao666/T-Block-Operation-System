// Package kafkasvr 从trpc-kafka复制而来,修改源码用于重置消费者的初始offset
package kafkasvr

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"golang.org/x/time/rate"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/metrics"
	"trpc.group/trpc-go/trpc-go/naming/selector"
	"trpc.group/trpc-go/trpc-go/transport"
	dsn "trpc.group/trpc-go/trpc-selector-dsn"
)

func init() {
	selector.Register("kafka", dsn.DefaultSelector)
}

// ContinueWithoutCommitError 是否在不 commit ack 的情况下，继续消费消息
// 场景：
// 在生产消息时，一个消息体可能会超过 kafka 的限制，因此会将原消息体拆分多个字节包，并封装成 kafka 的消息体，然后投递。
// 那么在消费消息时，就需要等所有的分包消息全部消费完毕方可开启业务逻辑处理。
// 当消费者的 Handle 方法或者 msg.ServerRspErr 返回该 error 时，则表示希望在不
// commit ack 的情况下，继续消费消息，而不会当做 error 处理。
var ContinueWithoutCommitError = &errs.Error{
	Type: errs.ErrorTypeBusiness,
	Code: errs.RetUnknown,
	Msg:  "Error:Continue to consume message without committing ack",
}

// Timeout 是 kafka 全局的 timeout 配置，默认 2s，用户有需要可自行修改。
var Timeout = 2 * time.Second

// IsCWCError Check if it is a ContinueWithoutCommitError error
// CWC:Continue Without Commit
func IsCWCError(err error) bool {
	return err == ContinueWithoutCommitError
}

var (
	serviceCloseError = errs.NewFrameError(errs.RetServerSystemErr, "kafka consumer service close")
	sessionCloseError = errs.NewFrameError(errs.RetServerSystemErr, "kafka consumer group session close")
	messageCloseError = errs.NewFrameError(errs.RetServerSystemErr, "kafka consumer group claim message close")
)

// Producer 封装 Producer 信息
type Producer struct {
	topic         string
	async         bool
	asyncProducer sarama.AsyncProducer
	syncProducer  sarama.SyncProducer
	trpcMeta      bool
	meta          codec.CommonMeta
}

// Close 封装 Close 动作
func (p *Producer) Close() error {
	if p.async {
		return p.asyncProducer.Close()
	}
	return p.syncProducer.Close()
}

// getPartitionMark 获取topic+partition唯一标识
func getPartitionMark(topic string, partition int32) string {
	return fmt.Sprintf("%s|%d", topic, partition)
}

// checkReady 判断所有分区是否ready
func checkReady(partitionReady []bool) bool {
	for _, val := range partitionReady {
		if !val {
			return false
		}
	}
	return true
}

// initOffset 初始化每个topic每个分区的offset
func initOffset(session sarama.ConsumerGroupSession, client sarama.Client,
	serviceName string, resetSec time.Duration) (bool, map[string]int32, error) {
	if resetSec <= 0 {
		return true, nil, nil
	}
	resetTime := time.Now().Add(-resetSec).UnixMilli()
	noData := true
	partitionIdxMap := make(map[string]int32)
	var idx int32 = 0
	for topic, partitions := range session.Claims() {
		for _, partition := range partitions {
			partitionIdxMap[getPartitionMark(topic, partition)] = idx
			idx = idx + 1
			offset, err := client.GetOffset(topic, partition, resetTime)
			if err != nil {
				return false, nil, err
			}
			if offset <= 0 {
				continue
			}
			noData = false
			session.ResetOffset(topic, partition, offset, "preload")
			session.MarkOffset(topic, partition, offset, "preload")
			log.Infof("kafka[%s]: success reset topic[%s] partition[%d] offset to [%.0f] second ago, new offset[%d] ",
				serviceName, topic, partition, resetSec.Seconds(), offset)
		}
	}
	return noData, partitionIdxMap, nil
}

// singleConsumerHandler 实现 sarama 消费者接口，逐条消费数据
type singleConsumerHandler struct {
	client        sarama.Client
	opts          *transport.ListenServeOptions
	ctx           context.Context
	retryMax      int           // 最大重试次数
	retryInterval time.Duration // 重试间隔
	trpcMeta      bool          // 是否传递 meta 信息
	limiter       *rate.Limiter // current limiter
	meta          codec.CommonMeta

	resetSec       time.Duration    // 重置offset到多少秒前
	isReady        bool             // 用于内部判断是否ready
	partitionIdx   map[string]int32 // 记录每个Topic的每个分区所属的idx
	partitionReady []bool           // 记录每个Topic的每个分区是否ready
	readyChain     chan struct{}    // 用于通知外部所有Topic是否ready
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (sh *singleConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	if sh.retryInterval == 0 {
		sh.retryInterval = time.Millisecond
	}
	if sh.isReady {
		return nil
	}
	ready, partitionIdxMap, err := initOffset(session, sh.client, sh.opts.ServiceName, sh.resetSec)
	if err != nil {
		return err
	}
	sh.isReady = ready
	sh.partitionIdx = partitionIdxMap
	sh.partitionReady = make([]bool, len(partitionIdxMap))
	if ready {
		sh.readyChain <- struct{}{}
	}
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (sh *singleConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (sh *singleConsumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		if err := sh.limiter.Wait(sh.ctx); err != nil {
			if err == sh.ctx.Err() {
				return serviceCloseError
			}
			return fmt.Errorf("kafka server transport: limiter error: %w", err)
		}
		select {
		case <-sh.ctx.Done(): // 判断服务是否结束
			return serviceCloseError
		case <-sess.Context().Done(): // 判断 session 是否结束
			return sessionCloseError
		case msg, ok := <-claim.Messages(): // 监听消息
			if !ok { // msg close  and return
				return messageCloseError
			}
			sh.retryConsumeAndMark(sess, msg, claim)
		}
	}
}

// RawSaramaContext 存放 sarama ConsumerGroupSession 和 ConsumerGroupClaim
// 导出此结构体是为了方便用户实现监控，提供的内容仅用于读，调用任何写方法属于未定义行为
type RawSaramaContext struct {
	Session sarama.ConsumerGroupSession
	Claim   sarama.ConsumerGroupClaim
}

type rawSaramaContextKey struct{}

func withRawSaramaContext(ctx context.Context, raw *RawSaramaContext) context.Context {
	return context.WithValue(ctx, rawSaramaContextKey{}, raw)
}

// GetRawSaramaContext 获取 sarama 原始上下文信息，包括 ConsumerGroupSession 和 ConsumerGroupClaim
// 获取到的上下文应该仅使用读方法，使用任何写方法都是未定义行为
func GetRawSaramaContext(ctx context.Context) (*RawSaramaContext, bool) {
	rawContext, ok := ctx.Value(rawSaramaContextKey{}).(*RawSaramaContext)
	return rawContext, ok
}

func (sh *singleConsumerHandler) retryConsumeAndMark(
	session sarama.ConsumerGroupSession,
	m *sarama.ConsumerMessage,
	claim sarama.ConsumerGroupClaim,
) {
	retryNum := 0
	// 如果需要重试，就等待一段时间后再次执行
	t := time.NewTimer(sh.retryInterval)
	defer t.Stop()
	for {
		ctx, trpcMsg := GenTRPCMessage(m, sh.opts.ServiceName, m.Topic, sh.trpcMeta, sh.meta)
		ctx = withRawSaramaContext(ctx, &RawSaramaContext{Session: session, Claim: claim})
		// 消费组没ready
		if !sh.isReady {
			partitionIdx := sh.partitionIdx[getPartitionMark(m.Topic, m.Partition)]
			// topic/partition没ready，消息时间和当前时间小于1s,则认为这个partition准备就绪
			if !sh.partitionReady[partitionIdx] && time.Now().Sub(m.Timestamp).Abs().Seconds() <= 1 {
				sh.partitionReady[partitionIdx] = true
				if checkReady(sh.partitionReady) {
					sh.isReady = true
					sh.readyChain <- struct{}{}
				}
			}
		}
		// 交给 trpc 框架处理
		_, err := sh.opts.Handler.Handle(ctx, nil)
		rspErr := trpcMsg.ServerRspErr()
		// 如果消费没有失败，跳出循环，mark message 并继续消费后续数据
		if err == nil && rspErr == nil {
			break
		}

		// 错误处理逻辑
		msgInfo := fmt.Sprintf("%s:%d:%d", m.Topic, m.Partition, m.Offset)
		// 如果是 IsCWCError 直接返回，不提交
		if IsCWCError(rspErr) {
			log.WarnContextf(ctx, "kafka consumer handle warn:%v, msg: %+v", rspErr, msgInfo)
			return
		}

		// 如果处理失败，写错误日志，重试次数 +1
		retryNum++
		log.ErrorContextf(ctx, "kafka consumer msg %s try time %d get fail:%v rspErr:%v, ", msgInfo, retryNum, err, rspErr)

		// 如果超过最大重试次数次数，结束循环 mark message 并继续消费后续数据
		if sh.retryMax != 0 && retryNum > sh.retryMax {
			consumerGroupName, _ := sh.meta[consumerGroupKey].(string)
			metrics.IncrCounter("trpc-kafka_"+consumerGroupName+"_DROP", 1)
			break
		}
		if !t.Stop() {
			select {
			case <-t.C:
			default:
			}
		}
		// Reset timer.
		t.Reset(sh.retryInterval)
		select {
		case <-sh.ctx.Done():
			return
		case <-session.Context().Done():
			return
		case <-t.C:
			// retry
		}
	}
	session.MarkMessage(m, "")
}

// batchConsumerHandler 批量消费的消费者
type batchConsumerHandler struct {
	opts          *transport.ListenServeOptions
	ctx           context.Context
	client        sarama.Client
	maxNum        int // 一批最大数量
	flushInterval time.Duration
	retryMax      int // 失败最大重试次数
	retryInterval time.Duration
	trpcMeta      bool          // 是否传递 meta 信息
	limiter       *rate.Limiter // current limiter
	meta          codec.CommonMeta

	resetSec       time.Duration    // 重置offset到多少秒前
	isReady        bool             // 用于内部判断是否ready
	partitionIdx   map[string]int32 // 记录每个Topic的每个分区所属的idx
	partitionReady []bool           // 记录每个Topic的每个分区是否ready
	readyChain     chan struct{}    // 用于通知外部所有Topic是否ready
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (bh *batchConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	if bh.retryInterval == 0 {
		bh.retryInterval = time.Millisecond
	}
	if bh.isReady {
		return nil
	}
	ready, partitionIdxMap, err := initOffset(session, bh.client, bh.opts.ServiceName, bh.resetSec)
	if err != nil {
		return err
	}
	bh.isReady = ready
	bh.partitionIdx = partitionIdxMap
	bh.partitionReady = make([]bool, len(partitionIdxMap))
	if ready {
		bh.readyChain <- struct{}{}
	}
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (bh *batchConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim 批量消费
// 当满足 maxNum 条消息时触发消费，刷新间隔到了也会触发消费，避免消息流量不均匀的情况下阻塞消费。
// 如果业务消费失败则整个批次重试，不支持只重试失败的消息
func (bh *batchConsumerHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	msgArray := make([]*sarama.ConsumerMessage, bh.maxNum)
	idx := 0

	ticker := time.NewTicker(bh.flushInterval)
	defer ticker.Stop()

	for {
		if err := bh.limiter.Wait(bh.ctx); err != nil {
			if err == bh.ctx.Err() {
				return serviceCloseError
			}
			return fmt.Errorf("kafka server transport: limiter error: %w", err)
		}
		select {
		case <-bh.ctx.Done(): // 判断服务是否结束
			return serviceCloseError
		case <-session.Context().Done(): // 判断 session 是否结束
			return sessionCloseError
		case msg, ok := <-claim.Messages():
			if !ok { // 如果 message close return
				return messageCloseError
			}

			msgArray[idx] = msg
			idx++

			if idx >= bh.maxNum { // 数据已经达到缓存最大值
				// 深度拷贝一份，否则下游异步处理的时候，msgArray 会被覆盖
				handleMsg := make([]*sarama.ConsumerMessage, len(msgArray))
				copy(handleMsg, msgArray)
				bh.retryConsumeAndMark(session, claim, handleMsg...)
				idx = 0
				resetTicker(ticker, bh.flushInterval)
			}
		case <-ticker.C:
			if idx > 0 {
				// 深度拷贝一份，否则下游异步处理的时候，msgArray 会被覆盖
				handleMsg := make([]*sarama.ConsumerMessage, idx)
				copy(handleMsg, msgArray[:idx])
				bh.retryConsumeAndMark(session, claim, handleMsg...)
				idx = 0
				resetTicker(ticker, bh.flushInterval)
			}
		}
	}
}

func (bh *batchConsumerHandler) retryConsumeAndMark(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
	msgs ...*sarama.ConsumerMessage,
) {
	retryNum := 0
	// 如果需要重试，就等待一段时间后再次执行
	t := time.NewTimer(bh.retryInterval)
	defer t.Stop()
	for {
		ctx, trpcMsg := GenTRPCMessage(msgs, bh.opts.ServiceName, msgs[0].Topic, bh.trpcMeta, bh.meta)
		ctx = withRawSaramaContext(ctx, &RawSaramaContext{Session: session, Claim: claim})
		// 消费组没ready
		m := msgs[0]
		if !bh.isReady {
			partitionIdx := bh.partitionIdx[getPartitionMark(m.Topic, m.Partition)]
			// topic/partition没ready，消息时间和当前时间小于1s,则认为这个partition准备就绪
			if !bh.partitionReady[partitionIdx] && time.Now().Sub(m.Timestamp).Abs().Seconds() <= 1 {
				bh.partitionReady[partitionIdx] = true
				if checkReady(bh.partitionReady) {
					bh.isReady = true
					bh.readyChain <- struct{}{}
				}
			}
		}
		// 交给 trpc 框架处理
		_, err := bh.opts.Handler.Handle(ctx, nil)
		rspErr := trpcMsg.ServerRspErr()
		// 如果消费没有失败，跳出循环，mark message 并继续消费后续数据
		if err == nil && rspErr == nil {
			break
		}

		// 如果处理失败，写错误日志，重试次数 +1
		retryNum++
		offset := make([]string, len(msgs))
		for i, v := range msgs {
			offset[i] = strconv.Itoa(int(v.Offset))
		}
		msg := msgs[0]
		info := fmt.Sprintf("topic: %s partition: %d offset: %s", msg.Topic, msg.Partition, strings.Join(offset, ","))
		log.ErrorContextf(ctx, "kafka consumer %s try number %d err: %v  msgErr: %v, ", info, retryNum, err, rspErr)

		// 如果超过最大重试次数次数，结束循环 mark message 并继续消费后续数据
		if bh.retryMax != 0 && retryNum > bh.retryMax {
			consumerGroupName, _ := bh.meta[consumerGroupKey].(string)
			metrics.IncrCounter("trpc-kafka_"+consumerGroupName+"_DROP", 1)
			break
		}

		if !t.Stop() {
			select {
			case <-t.C:
			default:
			}
		}
		// Reset timer.
		t.Reset(bh.retryInterval)
		select {
		case <-bh.ctx.Done():
			return
		case <-session.Context().Done():
			return
		case <-t.C:
			// retry
		}
	}
	session.MarkMessage(msgs[len(msgs)-1], "")
}

// GenTRPCMessage 生成新的 TRPC 消息，并保存 head，同时设置上下游服务名,
// 注意批量消费场景, 消费消息时会将首个消息的 trpc meta 填充到 ctx 中.
// 消费方可使用方法设置每条消息的 Head, 便于透传不同消息的真实 Head.
func GenTRPCMessage(reqHead interface{}, serviceName, topic string, trpcMeta bool, meta codec.CommonMeta) (
	context.Context, codec.Msg) {
	ctx, msg := codec.WithNewMessage(context.Background())
	msg.WithServerReqHead(reqHead)
	msg.WithCompressType(codec.CompressTypeNoop) // 不解压缩
	msg.WithCallerServiceName("trpc.kafka.noserver.noservice")
	msg.WithCallerMethod(topic)
	msg.WithCalleeServiceName(serviceName)
	msg.WithServerRPCName("/trpc.kafka.consumer.service/handle")
	msg.WithCalleeMethod(topic) // 修改被掉方法为 topic name
	if trpcMeta {
		m := getMessageHead(reqHead)
		setTRPCMeta(ctx, m) // 等同于 msg.WithServerMetaData(req.GetTransInfo())
		setDyeing(ctx, msg)
	}
	overrideCommonMeta(msg, meta) // 透传给其他插件如上报监控
	return ctx, msg
}

func overrideCommonMeta(msg codec.Msg, meta codec.CommonMeta) {
	mmeta := msg.CommonMeta()
	if mmeta == nil {
		mmeta = meta.Clone()
	} else {
		for k, v := range meta {
			mmeta[k] = v
		}
	}
	msg.WithCommonMeta(mmeta)
}

// setTRPCMeta sarama header 携带数据设置到 trpc meta
func setTRPCMeta(ctx context.Context, hs []*sarama.RecordHeader) {
	if hs == nil {
		return
	}
	for _, header := range hs {
		trpc.SetMetaData(ctx, string(header.Key), header.Value)
	}
}

// setDyeing 设置染色标记
func setDyeing(ctx context.Context, msg codec.Msg) {
	if dyeingValue := trpc.GetMetaData(ctx, trpc.DyeingKey); len(dyeingValue) != 0 {
		msg.WithDyeing(true)
		msg.WithDyeingKey(string(dyeingValue))
	}
}

// getMessageHead 从 reqHead 解析单个消息
func getMessageHead(reqHead interface{}) []*sarama.RecordHeader {
	switch reqHead := reqHead.(type) {
	case *sarama.ConsumerMessage:
		if reqHead != nil {
			return reqHead.Headers
		}
	case []*sarama.ConsumerMessage:
		// 批量消费的话，取第一个 message 的元数据
		if len(reqHead) > 0 && reqHead[0] != nil {
			return reqHead[0].Headers
		}
	}
	// default or slice empty
	return nil
}

func resetTicker(ticker *time.Ticker, d time.Duration) {
	ticker.Stop()
	select {
	case <-ticker.C:
	default:
	}
	ticker.Reset(d)
}
