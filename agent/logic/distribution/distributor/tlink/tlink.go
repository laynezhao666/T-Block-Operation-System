package tlink

import (
	"agent/utils/flog"
	"context"
	"encoding/json"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	model2 "agent/entity/model/data"
	utils2 "agent/logic/distribution/distributor/utils"
	collectorPb "trpcprotocol/collector"

	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/log"
)

const (
	tryNum               int           = 3
	distributorFilterKey string        = "tlink distributor"
	dataBusCallee        string        = "tbos.DataBus"   // DataBus服务的callee名称
	failedAddrCooldown   time.Duration = 60 * time.Second // 失败地址冷却时间
)

// failedAddrInfo 记录失败地址的信息
type failedAddrInfo struct {
	failTime time.Time // 失败时间
}

type tLinkDistributor struct {
	dataBusProxy  collectorPb.DataBusClientProxy
	targetAddrs   []string // 可用的后端地址列表，如 ["ip://1.2.3.4:30081", "ip://5.6.7.8:30081"]
	mu            sync.RWMutex
	failedAddrs   map[string]time.Time // 记录失败的地址及其失败时间，直接存储时间减少内存分配
	failedMu      sync.RWMutex         // 保护failedAddrs的锁
	lastCleanTime time.Time            // 上次清理时间
	cleanInterval time.Duration        // 清理间隔
}

var (
	tLinkDt   tLinkDistributor
	filterLog *flog.Filter
)

// Init 初始化
func Init() error {
	var err error
	if err = InitDistributors(); err != nil {
		return err
	}
	filterLog = flog.NewFilterLogger(time.Minute, log.GetDefaultLogger())
	return nil
}

// UnInit 反初始化
func UnInit() {
	UnInitDistributors()
}

// InitDistributors 初始化分发器
func InitDistributors() error {
	tLinkDt = tLinkDistributor{
		dataBusProxy:  collectorPb.NewDataBusClientProxy(),
		targetAddrs:   parseTargetAddrs(),
		failedAddrs:   make(map[string]time.Time),
		lastCleanTime: time.Now(),
		cleanInterval: 120 * time.Second, // 每120秒最多清理一次
	}
	return nil
}

// parseTargetAddrs 从trpc配置中解析DataBus服务的目标地址列表
// 将 "ip://1.2.3.4:30081,5.6.7.8:30081" 解析为 ["ip://1.2.3.4:30081", "ip://5.6.7.8:30081"]
func parseTargetAddrs() []string {
	addrs := make([]string, 0)

	// 遍历client配置，找到DataBus服务的target
	for _, svc := range trpc.GlobalConfig().Client.Service {
		if svc.Callee != dataBusCallee {
			continue
		}

		target := svc.Target
		if target == "" {
			continue
		}

		// 解析 ip://addr1,addr2,addr3 格式
		if strings.HasPrefix(target, "ip://") {
			addrPart := strings.TrimPrefix(target, "ip://")
			addrList := strings.Split(addrPart, ",")
			for _, addr := range addrList {
				addr = strings.TrimSpace(addr)
				if addr != "" {
					addrs = append(addrs, "ip://"+addr)
				}
			}
		}
		break
	}

	return addrs
}

// UnInitDistributors 反初始化分发器
func UnInitDistributors() {}

// TLinkDistributor 获取tlink分发器
func TLinkDistributor() *tLinkDistributor {
	return &tLinkDt
}

// getTargetAddrs 获取目标地址列表
func (t *tLinkDistributor) getTargetAddrs() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.targetAddrs
}

// markAddrFailed 标记地址失败
func (t *tLinkDistributor) markAddrFailed(addr string) {
	t.failedMu.Lock()
	t.failedAddrs[addr] = time.Now()
	t.failedMu.Unlock()
}

// clearAddrFailed 清除地址的失败记录
func (t *tLinkDistributor) clearAddrFailed(addr string) {
	t.failedMu.Lock()
	delete(t.failedAddrs, addr)
	t.failedMu.Unlock()
}

// tryCleanExpiredAddrs 尝试清理过期的失败地址记录（有频率限制）
func (t *tLinkDistributor) tryCleanExpiredAddrs() {
	now := time.Now()

	// 先用读锁检查是否需要清理
	t.failedMu.RLock()
	needClean := now.Sub(t.lastCleanTime) > t.cleanInterval && len(t.failedAddrs) > 0
	t.failedMu.RUnlock()

	if !needClean {
		return
	}

	// 需要清理，获取写锁
	t.failedMu.Lock()
	// 双重检查，避免多个goroutine重复清理
	if now.Sub(t.lastCleanTime) > t.cleanInterval {
		for addr, failTime := range t.failedAddrs {
			if now.Sub(failTime) > failedAddrCooldown {
				delete(t.failedAddrs, addr)
			}
		}
		t.lastCleanTime = now
	}
	t.failedMu.Unlock()
}

// getAvailableAddrs 获取可用的地址列表（排除冷却期内的地址）
// 一次性获取锁，减少锁竞争
func (t *tLinkDistributor) getAvailableAddrs(targetAddrs []string) []string {
	now := time.Now()
	available := make([]string, 0, len(targetAddrs))

	t.failedMu.RLock()
	for _, addr := range targetAddrs {
		failTime, exists := t.failedAddrs[addr]
		// 不存在失败记录，或者已过冷却期
		if !exists || now.Sub(failTime) > failedAddrCooldown {
			available = append(available, addr)
		}
	}
	t.failedMu.RUnlock()

	return available
}

// sendWithFailover 带故障转移的发送方法
// 首次使用默认proxy发送（利用其内置负载均衡算法），失败后切换到其他IP重试
// 会记录失败的地址，在冷却期内优先使用其他可用地址
func (t *tLinkDistributor) sendWithFailover(ctx context.Context, req *collectorPb.ReqSend, key string) error {
	var lastErr error
	// 尝试清理过期的失败地址记录（有频率限制，不会每次都执行）
	t.tryCleanExpiredAddrs()

	// 首次使用默认proxy发送，利用其内置的负载均衡算法
	_, lastErr = t.dataBusProxy.Send(ctx, req)
	if lastErr == nil {
		return nil
	}

	// 获取所有目标地址
	allAddrs := t.getTargetAddrs()
	// 首次发送失败，检查是否有可用的地址列表进行重试
	if len(allAddrs) <= 1 {
		// 只有一个或没有地址，无法进行故障转移，继续使用默认proxy重试
		for tryIdx := 1; tryIdx < tryNum; tryIdx++ {
			_, lastErr = t.dataBusProxy.Send(ctx, req)
			if lastErr == nil {
				filterLog.Debugf(distributorFilterKey+"debug", "key:%v try%d ok", key, tryIdx)
				return nil
			}
		}
		filterLog.Debugf(distributorFilterKey+"debug", "key:%v all retry fail", key)
		return lastErr
	}

	// 获取可用地址（延迟到真正需要重试时才获取）
	availableAddrs := t.getAvailableAddrs(allAddrs)
	filterLog.Debugf(distributorFilterKey+"debug", "key:%v failover start, all:%d avail:%d",
		key, len(allAddrs), len(availableAddrs))

	// 有多个地址，进行故障转移重试
	// 优先使用可用地址（不在冷却期的地址），如果可用地址用完再使用冷却期内的地址
	// 重试次数确保能遍历所有地址
	addrCount := len(allAddrs)
	retryCount := tryNum - 1 // 已经尝试过1次，还需要重试的次数
	if addrCount > retryCount {
		retryCount = addrCount
	}

	// 记录本次请求中已尝试过的地址
	triedAddrs := make(map[string]struct{}, addrCount)

	// 使用随机起始位置，确保流量均匀分布
	startIdx := 0
	if len(availableAddrs) > 0 {
		startIdx = rand.Intn(len(availableAddrs))
	}

	tryIdx := 0
	// 第一阶段：优先尝试可用地址（不在冷却期的地址）
	for i := 0; i < len(availableAddrs) && tryIdx < retryCount; i++ {
		targetAddr := availableAddrs[(startIdx+i)%len(availableAddrs)]
		triedAddrs[targetAddr] = struct{}{}

		_, lastErr = t.dataBusProxy.Send(ctx, req, client.WithTarget(targetAddr))
		if lastErr == nil {
			filterLog.Debugf(distributorFilterKey+"debug", "key:%v try%d ok target:%v",
				key, tryIdx+1, targetAddr)
			return nil
		}

		// 标记该地址失败
		t.markAddrFailed(targetAddr)
		filterLog.Debugf(distributorFilterKey+"debug", "key:%v try%d fail target:%v err:%v",
			key, tryIdx+1, targetAddr, lastErr)
		tryIdx++
	}

	// 第二阶段：如果可用地址都失败了，尝试冷却期内的地址（可能已经恢复）
	if tryIdx < retryCount {
		filterLog.Debugf(distributorFilterKey+"debug", "key:%v try cooldown addrs", key)
	}
	for i := 0; i < addrCount && tryIdx < retryCount; i++ {
		targetAddr := allAddrs[i]
		if _, tried := triedAddrs[targetAddr]; tried {
			continue // 跳过已尝试的地址
		}
		triedAddrs[targetAddr] = struct{}{}

		_, lastErr = t.dataBusProxy.Send(ctx, req, client.WithTarget(targetAddr))
		if lastErr == nil {
			// 成功了，清除该地址的失败记录
			t.clearAddrFailed(targetAddr)
			filterLog.Debugf(distributorFilterKey+"debug", "key:%v try%d(cd) ok target:%v",
				key, tryIdx+1, targetAddr)
			return nil
		}

		// 更新该地址的失败时间
		t.markAddrFailed(targetAddr)
		filterLog.Debugf(distributorFilterKey+"debug", "key:%v try%d(cd) fail target:%v err:%v",
			key, tryIdx+1, targetAddr, lastErr)
		tryIdx++
	}
	return lastErr
}

// Distribute 分发
func (t *tLinkDistributor) Distribute(data *model2.DataUnit, args ...interface{}) {
	if t == nil || data == nil || len(data.Points) == 0 {
		return
	}
	sendTime, interval, dataType := utils2.GetSendTimeAndInterval(args)
	mozuID := utils2.GetMozuID(args)

	kData, kafkaDataList, err := utils2.ToKafkaData(data, interval, false)
	if err != nil {
		filterLog.Errorf(distributorFilterKey, "tlink Distributor.Distribute %+v error: %+v", data, err)
		return
	}
	if len(kData.Points) == 0 && len(kData.VirtualPoints) == 0 {
		return
	}

	utils2.DebugRecord(kData)
	messageLen := 0
	var (
		wg       sync.WaitGroup
		sendOk   atomic.Int64
		sendFail atomic.Int64
	)
	for _, kafkaData := range kafkaDataList {
		b, err := json.Marshal(kafkaData)
		if err != nil {
			filterLog.Errorf(distributorFilterKey, "tlinkDistributor.Distribute Marshal %+v error: %+v", kafkaData, err)
			continue
		}
		// 计算当前消息的实际测点数
		n := len(kafkaData.Points) + len(kafkaData.VirtualPoints)
		key := utils2.GetMessageKey(data.DeviceGid, sendTime.Unix(), interval, dataType, n)
		messageLen += len(b)
		wg.Add(1)
		go func(wg *sync.WaitGroup, key []byte, value []byte, pointCount int64) {
			defer wg.Done()
			req := &collectorPb.ReqSend{
				Key:   key,
				Value: value,
			}

			// 使用带故障转移的发送方法
			ctx := context.Background()
			sendErr := t.sendWithFailover(ctx, req, string(key))
			if sendErr != nil {
				filterLog.Errorf(distributorFilterKey+"error", "key: %v, tlink Distribute Send error: %+v", string(key), sendErr)
				sendFail.Add(pointCount)
			} else {
				sendOk.Add(pointCount)
			}
		}(&wg, []byte(key), b, int64(n))
	}

	shouldLog := utils2.GetLogger(data.DeviceGid).Insert(interval)
	if shouldLog {
		kData.Log("tlink", data.DeviceGid, interval)
	}
	wg.Wait()
	kData.ReportWithStats("tlink", interval, mozuID, int(sendOk.Load()), int(sendFail.Load()), dataType)
}
