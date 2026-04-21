package data

import (
	"context"
	"dac/entity/config"
	"dac/entity/consts"
	"dac/entity/model/rt"
	"encoding/json"
	"fmt"
	"github.com/samber/lo"
	"strings"
	"sync"
	"time"
	"trpc.group/trpc-go/trpc-go/log"

	"dac/entity/model/db"

	pointpb "dac/repo/pb/tcommon_point_data"
	collectorPb "trpcprotocol/collector"
)

var (
	tbosClient = collectorPb.NewDataBusClientProxy()
	tryNum     = 3

	// 标准点配置缓存，key为门禁控制器code，从cmdb获取，用于测点转化上报tbos
	stdPointsCache = make(map[string]*rt.StdInstancePointsInfo)
	stdPointsMutex = sync.RWMutex{}

	// 标准点配置版本缓存，key为门禁控制器code，用于判断是否重新查询配置更新缓存
	cmdbVersionCache = make(map[string]*rt.ConfigVersion)
	cmdbVersionMutex = sync.RWMutex{}
)

// writerImpl TBOS数据上报的实现结构体
type writerImpl struct{}

// StartRefreshTBOSCacheLoop 启动CMDB版本和标准点配置缓存刷新循环任务
func StartRefreshTBOSCacheLoop(ctx context.Context) {
	go refreshTBOSCacheLoop(ctx)
	go cleanUpTBOSCacheLoop(ctx)
}

// cleanUpTBOSCacheLoop 定期清理TBOS缓存
func cleanUpTBOSCacheLoop(ctx context.Context) {
	for {
		// 每天凌晨2点刷新缓存
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, now.Location())
		duration := next.Sub(now)

		select {
		case <-time.After(duration):
			clearTBOSCache()
		case <-ctx.Done():
			log.Info("stop clean up TBOS cache loop.")
			return
		}
	}
}

// refreshTBOSCacheLoop CMDB版本和标准点配置缓存刷新循环任务
func refreshTBOSCacheLoop(ctx context.Context) {
	for {
		refreshTBOSCache(ctx)
		select {
		case <-time.After(5 * time.Minute): // 每5分钟执行一次
			break
		case <-ctx.Done():
			log.Info("stop refresh TBOS cache loop.")
			return
		}
	}
}

// refreshTBOSCache 刷新缓存
func refreshTBOSCache(ctx context.Context) {
	// 获取所有缓存中的设备code
	codes := getAllCachedCodes()
	if len(codes) == 0 {
		return
	}

	log.Debugf("开始刷新 %d 个设备的缓存", len(codes))

	// 批量获取版本信息
	versionMap, err := getCmdbVersionBatch(ctx, codes)
	if err != nil {
		log.Warnf("批量获取版本信息失败: %v", err)
		return
	}

	// 检查每个设备是否需要更新
	for _, code := range codes {
		currentVersion, exists := versionMap[code]
		if !exists {
			continue
		}

		cachedVersion := getCachedVersion(code)
		if needUpdateCache(cachedVersion, currentVersion) {
			log.Infof("设备 %s cmdb版本变更，更新缓存", code)
			updateStdPointsCache(ctx, code, currentVersion)
		}
	}
}

// getAllCachedCodes 获取所有缓存中的设备code
func getAllCachedCodes() []string {
	stdPointsMutex.RLock()
	defer stdPointsMutex.RUnlock()

	codes := make([]string, 0, len(stdPointsCache))
	for code := range stdPointsCache {
		codes = append(codes, code)
	}
	return codes
}

// SetTBOSPointsWithExtends 门禁采集测点转标准测点上报tbos
func (w *writerImpl) SetTBOSPointsWithExtends(ctx context.Context, timestamp int64, kind pointpb.DataKind, extends map[string]string, code string, deviceID db.GIDType, points []*pointpb.Point) error {
	// 增加tbos上报
	chunkSize := 500
	chunks := lo.Chunk(points, chunkSize)

	// 获取标准点配置（使用缓存）
	stdPoints, err := getStdPointsConfigWithCache(ctx, code)
	if err != nil {
		log.Errorf("获取标准点配置失败，code：%s，error：%s", code, err)
		return err
	}
	if stdPoints == nil || len(*stdPoints) == 0 {
		log.Debugf("设备 %s 没有标准点配置，跳过上报tbos", code)
		return nil
	}

	for _, chunk := range chunks {
		// 转化测点数据，上报的tbos
		key, value := convertToTbosFormat(*stdPoints, timestamp, kind, extends, code, deviceID, chunk)
		if config.C.Debug {
			log.Infof("查询到门禁设备的标准点配置为：%v，本次上报的pbPoint为：%v", stdPoints, chunk)
		}
		// 如果没有有效的标准点，跳过本次上报
		if key == nil || value == nil {
			log.Debugf("设备 %s 没有有效的标准点映射，跳过上报tbos", code)
			continue
		}
		tbosReq := &collectorPb.ReqSend{
			Key:   key,
			Value: value,
		}
		for i := 0; i < tryNum; i++ { // 重试3次
			if _, err = tbosClient.Send(context.Background(), tbosReq); err == nil {
				break
			}
			log.Warnf("第%d次重试上报门禁测点到tbos, key: %v, Send error: %+v", i, key, err)
		}
	}
	return nil
}

// convertToTbosFormat 转换为tbos格式测点
func convertToTbosFormat(stdPoints rt.StdInstancePointsInfo, timestamp int64, kind pointpb.DataKind, extends map[string]string, code string, deviceID db.GIDType, points []*pointpb.Point) ([]byte, []byte) {
	if timestamp <= 0 {
		timestamp = time.Now().UnixMilli()
	}
	// 创建标准点映射表：key为采集点ID，value为标准点point_key
	stdPointMap := make(map[string]string)
	for _, point := range stdPoints {
		// 解析expression_map获取采集点ID
		exprParts := strings.Split(point.Mapping, "=")
		if len(exprParts) >= 2 {
			// 获取等号后面的采集点ID（去掉可能的分号）
			sourcePointID := strings.TrimSuffix(exprParts[1], ";")
			stdPointMap[sourcePointID] = point.PointKey
		}
	}

	var interval int32
	if kind == pointpb.DataKind_Period {
		interval = rt.PointIntervalPeriod
	} else if kind == pointpb.DataKind_Change {
		interval = rt.PointIntervalChange
	}

	key := rt.MsgKey{
		Timestamp: timestamp / 1000,
		Interval:  interval,
		Type:      rt.PointTypeStd,
		MozuID:    extends[consts.PointMessageMozuKey],
		DeviceGiD: string(deviceID),
		WorkerID:  code,
		//Seq: ,
		//BalancerKey: ,
		PubMs: time.Now().UnixMilli(),
	}

	var msgPoints []*rt.MsgPoint
	for _, p := range points {
		// 只有找到标准点映射才上报，否则跳过
		if mappedKey, ok := stdPointMap[p.Id]; ok {
			msgPoints = append(msgPoints, &rt.MsgPoint{
				I: mappedKey,
				V: fmt.Sprintf("%v", p.Value),
				Q: fmt.Sprintf("%d", p.Quality),
				T: fmt.Sprintf("%d", p.Timestamp),
			})
		}
	}

	// 如果没有任何有效的标准点，返回nil
	if len(msgPoints) == 0 {
		return nil, nil
	}

	value := rt.MsgValue{
		Interval: int64(1),
		Points:   msgPoints,
	}
	if config.C.Debug {
		log.Infof("本次上报的标准测点，key：%s，value：%s", key, value)
	}

	keyBytes, _ := json.Marshal(key)
	valueBytes, _ := json.Marshal(value)
	return keyBytes, valueBytes
}

// getStdPointsConfigWithCache 从缓存获取标准点配置
func getStdPointsConfigWithCache(ctx context.Context, code string) (*rt.StdInstancePointsInfo, error) {
	// 先尝试从缓存获取
	if stdPoints, exists := getStdPointsFromCache(code); exists {
		if config.C.Debug {
			log.Infof("从缓存获取到门禁设备 %s 的标准点配置", code)
		}
		return stdPoints, nil
	}

	// 缓存中没有，从远程获取
	devices := []string{code}
	params, err := json.Marshal(devices)
	if err != nil {
		log.Errorf("门禁控制器code序列化失败:%v", devices)
		return nil, err
	}

	stdReq := &collectorPb.ReqFetchConfig{
		Params:    params,
		FetchType: collectorPb.ReqFetchConfig_FETCH_STD_POINTS,
	}
	configClient := collectorPb.NewConfigBusClientProxy()
	rsp, err := configClient.FetchConfig(ctx, stdReq)
	if err != nil {
		log.Errorf("查询idc-tbos-cmdb标准测点映射失败，device：%s，err：%s", devices, err)
		return nil, err
	}

	data := rsp.GetData()
	var configMap map[string]any
	err = json.Unmarshal(data, &configMap)
	if err != nil {
		log.Errorf("idc-tbos-cmdb标准测点映射反序列化失败, data：%v，error：%s", data, err)
		return nil, err
	}

	// 解析标准点配置
	stdPoints, err := ParseStdPointConfigMap(configMap)
	if err != nil {
		log.Errorf("标准点配置转化失败，configMap：%v，error：%s", configMap, err)
		return nil, err
	}

	// 存入缓存
	setStdPointsToCache(code, stdPoints)

	if config.C.Debug {
		log.Infof("从远程获取并缓存门禁设备 %s 的标准点配置", code)
	}

	return stdPoints, nil
}

// ParseStdPointConfigMap 从configMap解析标准测点配置
func ParseStdPointConfigMap(configMap map[string]any) (*rt.StdInstancePointsInfo, error) {
	stdPoints := new(rt.StdInstancePointsInfo)
	if len(configMap) == 0 {
		log.Warnf("std points config not exist")
		return stdPoints, nil
	}
	var allDevicePoints []any
	for _, info := range configMap {
		var conf rt.PointConfig
		b, err := json.Marshal(info)
		if err != nil {
			log.Warnf("marshal std point config err: %v", err)
			continue
		}
		err = json.Unmarshal(b, &conf)
		if err != nil {
			log.Warnf("unmarshal std point config err: %v", err)
			continue
		}
		allDevicePoints = append(allDevicePoints, conf.DevicePoints...)
	}
	stdJson, _ := json.Marshal(allDevicePoints)

	err := json.Unmarshal([]byte(stdJson), stdPoints)
	if err != nil {
		log.Error("std point config: unmarshal fail:%v", err)
		return nil, err
	}
	return stdPoints, nil
}

// GetCmdbVersionBatch 根据门控器名称批量获取cmdb版本
func getCmdbVersionBatch(ctx context.Context, codes []string) (map[string]*rt.ConfigVersion, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	params, err := json.Marshal(codes)
	if err != nil {
		return nil, err
	}
	req := &collectorPb.ReqFetchConfig{
		Params:    params,
		FetchType: collectorPb.ReqFetchConfig_FETCH_CONFIG_MODIFY_TIME,
	}
	configClient := collectorPb.NewConfigBusClientProxy()
	rsp, err := configClient.FetchConfig(ctx, req)
	if err != nil {
		log.Errorf("查询CMDB版本信息失败，device：%s，err：%s", params, err)
		return nil, err
	}

	data := rsp.GetData()
	var configMap map[string]any
	err = json.Unmarshal(data, &configMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal cmdb version config map err: %v", err)
	}

	return ParseCmdbVersionConfigMap(configMap)
}

// ParseCmdbVersionConfigMap 从configMap解析cmdb版本配置
func ParseCmdbVersionConfigMap(configMap map[string]any) (map[string]*rt.ConfigVersion, error) {
	if len(configMap) == 0 {
		log.Warnf("cmdb version config not exist")
		return nil, nil
	}
	versionMap := make(map[string]*rt.ConfigVersion, len(configMap))
	for k, v := range configMap {
		version := &rt.ConfigVersion{}
		b, err := json.Marshal(v)
		if err != nil {
			log.Warnf("marshal cmdb version device config err: %v", err)
			continue
		}
		err = json.Unmarshal(b, version)
		if err != nil {
			log.Warnf("unmarshal cmdb version device config err: %v", err)
			continue
		}
		versionMap[k] = version
	}
	return versionMap, nil
}

// getCachedVersion 获取缓存中的版本
func getCachedVersion(code string) *rt.ConfigVersion {
	cmdbVersionMutex.RLock()
	defer cmdbVersionMutex.RUnlock()
	return cmdbVersionCache[code]
}

// needUpdateCache 判断是否需要更新缓存
func needUpdateCache(cached, current *rt.ConfigVersion) bool {
	if cached == nil {
		return true
	}
	if current == nil {
		return false
	}
	return cached.Point != current.Point
}

// updateStdPointsCache 更新标准点配置缓存
func updateStdPointsCache(ctx context.Context, code string, version *rt.ConfigVersion) {
	devices := []string{code}
	params, err := json.Marshal(devices)
	if err != nil {
		log.Warnf("序列化设备code失败: %v", err)
		return
	}

	stdReq := &collectorPb.ReqFetchConfig{
		Params:    params,
		FetchType: collectorPb.ReqFetchConfig_FETCH_STD_POINTS,
	}
	configClient := collectorPb.NewConfigBusClientProxy()
	rsp, err := configClient.FetchConfig(ctx, stdReq)
	if err != nil {
		log.Warnf("查询标准点配置失败: %v", err)
		return
	}

	data := rsp.GetData()
	var configMap map[string]any
	err = json.Unmarshal(data, &configMap)
	if err != nil {
		log.Warnf("反序列化标准点配置失败: %v", err)
		return
	}

	stdPoints, err := ParseStdPointConfigMap(configMap)
	if err != nil {
		log.Warnf("解析标准点配置失败: %v", err)
		return
	}

	// 更新缓存
	setStdPointsToCache(code, stdPoints)
	setCachedVersion(code, version)
	log.Infof("已更新设备 %s 的标准点配置缓存", code)
}

// setCachedVersion 设置cmdb版本到缓存
func setCachedVersion(code string, version *rt.ConfigVersion) {
	cmdbVersionMutex.Lock()
	defer cmdbVersionMutex.Unlock()
	cmdbVersionCache[code] = version
}

// getStdPointsFromCache 从缓存获取标准点配置
func getStdPointsFromCache(code string) (*rt.StdInstancePointsInfo, bool) {
	stdPointsMutex.RLock()
	defer stdPointsMutex.RUnlock()

	stdPoints, exists := stdPointsCache[code]
	return stdPoints, exists
}

// setStdPointsToCache 设置标准点配置到缓存
func setStdPointsToCache(code string, stdPoints *rt.StdInstancePointsInfo) {
	stdPointsMutex.Lock()
	defer stdPointsMutex.Unlock()

	stdPointsCache[code] = stdPoints
}

// clearTBOSCache 清理TBOS缓存
func clearTBOSCache() {
	// 清理标准点配置缓存
	stdPointsMutex.Lock()
	stdPointsCacheCount := len(stdPointsCache)
	stdPointsCache = make(map[string]*rt.StdInstancePointsInfo)
	stdPointsMutex.Unlock()

	// 清理版本缓存
	cmdbVersionMutex.Lock()
	versionCacheCount := len(cmdbVersionCache)
	cmdbVersionCache = make(map[string]*rt.ConfigVersion)
	cmdbVersionMutex.Unlock()

	log.Infof("定期清理TBOS缓存完成: 清理标准点配置 %d 个, CMDB版本信息 %d 个", stdPointsCacheCount, versionCacheCount)
}
