package collectors_config

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"collector/entity/errcode"
	"collector/logic/bus/collectors_config/cmdb"
	"collector/logic/bus/collectors_config/localfile"
	"collector/utils"

	"collector/entity/collectors"
	"collector/repo/report"

	"etrpc-go/log"
	pb "trpcprotocol/collector"

	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/metrics"
)

const (
	primaryTimeRatio float64       = float64(0.5)
	unlimitedTimeout time.Duration = time.Duration(1 * time.Second)
	handleType       string        = "fetch config"
)

var (
	fetchers = make([]IConfigFetcher, 0)
)

var (
	fixedDimensions = []*metrics.Dimension{
		{
			Name:  report.HandleTypeDimension,
			Value: "fetch config",
		},
	}
)

// Init 初始化
func Init() {
	// 需要创建的目录列表
	dirs := []struct {
		path        string
		description string
	}{
		{collectors.CollectDevicesConfigDir, "collect devices config"},
		{collectors.CollectTemplatesConfigDir, "collect templates config"},
		{collectors.StdPointsConfigDir, "collect points config"},
		{collectors.ConfigModifyTimeDir, "config modify time"},
		{collectors.StdDevicesConfigDir, "std devices config"},
	}

	// 批量创建目录
	for _, dir := range dirs {
		if !utils.IsExist(dir.path) {
			if err := utils.CreateDir(dir.path); err != nil {
				log.Warnf("create %s dir fail: %v", dir.description, err)
			}
		}
	}
	Register(cmdb.Fetcher(), localfile.Fetcher())
	// Register(localfile.Fetcher(), cmdb.Fetcher())
	// Register(localfile.Fetcher())
	log.Info("collect config fetchers init done")
}

// Register 注册配置获取器
func Register(f ...IConfigFetcher) {
	fetchers = append(fetchers, f...)
}

// FetchConfigHandle 处理关于配置的请求
// 会按照注册顺序选取获取器进行配置的获取，失败则尝试下一个
func FetchConfigHandle(ctx context.Context, req *pb.ReqFetchConfig) ([]byte, error) {
	defer utils.HandlePanic("fetch_config")
	startTime := time.Now().UnixMilli()
	handleCnt := 1
	handleFailCnt := 0
	defer func() {
		endTime := time.Now().UnixMilli()
		t := float64(endTime - startTime)
		additionalDimensions := []*metrics.Dimension{
			{
				Name:  report.UpstreamIpDimension,
				Value: utils.GetUpstreamIp(ctx),
			},
		}
		reportHandleMetric(float64(handleCnt), float64(handleFailCnt), t, additionalDimensions)
	}()
	params := []string{}
	// 此处使用jsoniter，并且请求时不带参数时（如curl 127.0.0.1:8080/ConfigBus/FetchConfig）
	// 会返回curl: (8) Nul byte in header
	// 使用json.Unmarshal时，会返回{"code":270202,
	// "message":"unmarshal params error: unexpected end of JSON input",
	// "data":null,"trace_id":"4e746037497bcc3ac318fa961ef9ca4c"}
	err := json.Unmarshal(req.GetParams(), &params)
	if err != nil {
		handleFailCnt += 1
		return nil, errs.New(errcode.ErrRequestContentMissed, fmt.Sprintf("unmarshal params error: %v", err))
	}
	if len(params) == 0 {
		handleFailCnt += 1
		return nil, errs.New(errcode.ErrRequestContentMissed, "no params in request")
	}
	b, err := FetchConfig(ctx, req.FetchType, params)
	if err != nil {
		handleFailCnt += 1
		return nil, err
	}
	return b, nil
}

// FetchConfig 获取配置
func FetchConfig(ctx context.Context, fetchType pb.ReqFetchConfig_FetchType, params []string) ([]byte, error) {
	if len(fetchers) == 0 {
		return nil, errs.New(errcode.ErrFetchConfigFail, "no fetcher")
	}

	// 获取对应的fetch函数
	fetchFunc := getFetcherFunc(fetchType)
	if fetchFunc == nil {
		return nil, errs.New(errcode.ErrFetchConfigFail, fmt.Sprintf("unknown fetch type <%v>", fetchType))
	}

	timeoutList := getTimeoutList(ctx)
	var errors []error

	// 遍历所有fetcher尝试获取配置
	for i, f := range fetchers {
		reportFetchConfigMetric(1, 0, fmt.Sprintf("%v", fetchType), f.Name())
		value, err := fetchConfigWithTimeout(ctx, timeoutList[i], params, fetchFunc(f))
		if err != nil {
			reportFetchConfigMetric(0, 1, fmt.Sprintf("%v", fetchType), f.Name())
			log.WarnContextf(ctx, "fetchtype <%v>  %v fetch error: %v, fetcher: <%v>", fetchType, params, err, f.Name())
			errors = append(errors, err)
			continue
		}
		log.InfoContextf(ctx, "fetchtype <%v> %v fetch success, fetcher: <%v>", fetchType, params, f.Name())
		return value, nil
	}

	errMsg := fmt.Sprintf("fetchtype <%v>, keys %v, fetch fail: %v", fetchType, params, errors)
	log.ErrorContextf(ctx, errMsg)
	return nil, errs.New(errcode.ErrFetchConfigFail, errMsg)
}

// getFetcherFunc 根据fetchType返回对应的fetcher函数
func getFetcherFunc(fetchType pb.ReqFetchConfig_FetchType) func(IConfigFetcher) func([]string) ([]byte, error) {
	switch fetchType {
	case pb.ReqFetchConfig_FETCH_COLLECTOR_DEVICES:
		return func(f IConfigFetcher) func([]string) ([]byte, error) { return f.FetchCollectDevices }
	case pb.ReqFetchConfig_FETCH_COLLECTOR_TEMPLATES:
		return func(f IConfigFetcher) func([]string) ([]byte, error) { return f.FetchCollectTemplates }
	case pb.ReqFetchConfig_FETCH_STD_POINTS:
		return func(f IConfigFetcher) func([]string) ([]byte, error) { return f.FetchStdPoints }
	case pb.ReqFetchConfig_FETCH_CONFIG_MODIFY_TIME:
		return func(f IConfigFetcher) func([]string) ([]byte, error) { return f.FetchConfigModifyTime }
	case pb.ReqFetchConfig_FETCH_STD_DEVICES:
		return func(f IConfigFetcher) func([]string) ([]byte, error) { return f.FetchStdDevices }
	default:
		return nil
	}
}

// 根据context获取超时时间列表
func getTimeoutList(ctx context.Context) []time.Duration {
	timeoutList := make([]time.Duration, len(fetchers))
	var totalTimeout, primaryTimeout, secondaryTimeout time.Duration
	deadline, ok := ctx.Deadline()
	if ok {
		totalTimeout = time.Until(deadline)
		primaryTimeout = time.Duration(float64(totalTimeout) * primaryTimeRatio)
		secondaryTimeout = time.Duration(float64(totalTimeout-primaryTimeout) / float64(len(fetchers)-1))
		log.DebugContextf(ctx, "remaining time: %v, time for priority fetcher: %v", totalTimeout, primaryTimeout)
		timeoutList[0] = primaryTimeout
		for i := 1; i < len(fetchers); i++ {
			timeoutList[i] = secondaryTimeout
		}
	} else {
		for i := 0; i < len(fetchers); i++ {
			timeoutList[i] = unlimitedTimeout
		}
	}
	return timeoutList
}

func fetchConfigWithTimeout(ctx context.Context, timeout time.Duration, params []string, fetchFunction func(deviceNumbers []string) ([]byte, error)) ([]byte, error) {
	withTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	valueCh := make(chan []byte, 1)
	errCh := make(chan error, 1)
	go func(valueCh chan []byte, errCh chan error) {
		value, err := fetchFunction(params)
		if err != nil {
			errCh <- err
		} else {
			valueCh <- value
		}
	}(valueCh, errCh)
	select {
	// 优先检查父ctx的取消
	case <-ctx.Done():
		return nil, fmt.Errorf("ctx %v", ctx.Err())
	case <-withTimeout.Done():
		return nil, fmt.Errorf("withtimeout %v", withTimeout.Err())
	case err := <-errCh:
		return nil, err
	case value := <-valueCh:
		return value, nil
	}
}

func reportHandleMetric(handleCnt, handleFailCnt, latency float64, additionalDimensions []*metrics.Dimension) {
	dimensions := append(fixedDimensions, additionalDimensions...)
	report.HandleCnt(dimensions, handleCnt)
	report.HandleFailCnt(dimensions, handleFailCnt)
	report.HandleLatency(dimensions, latency)
}

func reportFetchConfigMetric(fetchConfigCnt, fetchConfigFailCnt float64, fetchType, fetcherName string) {
	dimensions := []*metrics.Dimension{
		{
			Name:  report.FetchTypeDimension,
			Value: fetchType,
		},
		{
			Name:  report.FetcherNameDimension,
			Value: fetcherName,
		},
	}
	report.FetchConfigCnt(dimensions, fetchConfigCnt)
	report.FetchConfigCnt(dimensions, fetchConfigFailCnt)
}
