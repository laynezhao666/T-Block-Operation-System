package http

import (
	"context"
	"sync"
	"time"

	"trpc.group/trpc-go/trpc-go/client"

	"agent/entity/config"
	"agent/utils/flog"
	"agent/utils/osal"

	"etrpc-go/util/httputil"

	"trpc.group/trpc-go/trpc-go/log"

	model2 "agent/entity/model/data"
	utils2 "agent/logic/distribution/distributor/utils"
)

const (
	tryNum               int    = 3
	distributorFilterKey string = "http distributor"
)

type httpDistributor struct {
	clients   map[string]*httpClient
	mutex     sync.RWMutex
	whitelist osal.Set[string]
}

var (
	httpDt    httpDistributor
	filterLog *flog.Filter
)

// rspHttpDistribute http上报响应参数
type rspHttpDistribute struct {
	code    int32
	message string
}

// Init 初始化
func Init() error {
	var err error
	if err = InitDistributors(); err != nil {
		return err
	}
	return nil
}

// UnInit 退出
func UnInit() {
	UnInitDistributors()
}

// InitDistributors 初始化分发器
func InitDistributors() error {
	httpDt = httpDistributor{
		clients:   make(map[string]*httpClient),
		mutex:     sync.RWMutex{},
		whitelist: osal.NewSet[string](),
	}
	filterLog = flog.NewFilterLogger(time.Minute, log.GetDefaultLogger())
	whitelist := config.GetRB().Distributor.Http.NorthWhitelist
	for _, p := range whitelist {
		httpDt.whitelist.Add(p)
	}
	log.Infof("north whitelist %v", httpDt.whitelist)
	return nil
}

// UnInitDistributors 退出分发器
func UnInitDistributors() {}

// HttpDistributor 获取http分发器
func HttpDistributor() *httpDistributor {
	return &httpDt
}

// Distribute 分发数据
func (t *httpDistributor) Distribute(data *model2.DataUnit, args ...interface{}) {
	if t == nil || data == nil || len(data.Points) == 0 || len(t.clients) == 0 {
		return
	}
	_, interval := utils2.GetSendTimeAndInterval(args)

	kData, kafkaDataList, err := utils2.ToKafkaData(data, interval, true)
	if err != nil {
		filterLog.Errorf(distributorFilterKey, "http Distributor.Distribute %+v error: %+v", data, err)
		return
	}
	if len(kData.Points) == 0 && len(kData.VirtualPoints) == 0 {
		return
	}

	utils2.DebugRecord(kData)
	var wg sync.WaitGroup
	for _, kafkaData := range kafkaDataList {
		t.mutex.RLock()
		defer t.mutex.RUnlock()
		for _, cli := range t.clients {
			wg.Add(1)
			go func(wg *sync.WaitGroup, c *httpClient, kafkaData *utils2.KafkaData) {
				defer wg.Done()
				kafkaData.Box.ClientId = c.config.ClientId
				errors := []error{}
				for tryIdx := 0; tryIdx < tryNum; tryIdx++ {
					err := httputil.PostJson(context.Background(), c.config.Target, nil, kafkaData, nil,
						client.WithTLS(
							"",     // 证书路径, 留空
							"",     // 私钥路径, 留空
							"none", // CA 证书, 两边都不认证
							"",     // server name, 留空
						))
					if err == nil {
						log.Debugf("http Distribute success")
						return
					}
					errors = append(errors, err)
					filterLog.Debugf(distributorFilterKey+"debug", "try idx %v,http Distribute Send error: %+v", tryIdx, err)
				}
				filterLog.Errorf(distributorFilterKey+"error", "http Distribute Send error: %+v", errors)

			}(&wg, cli, kafkaData.Copy())
		}
	}

	shouldLog := utils2.GetLogger(data.DeviceGid).Insert(interval)
	if shouldLog {
		kData.Log("http", data.DeviceGid, interval)
	}
	kData.Report("http", interval)
	wg.Wait()
}
