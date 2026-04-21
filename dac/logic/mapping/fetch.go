package mapping

import (
	"context"
	"dac/entity/config"
	"dac/entity/model/db"
	"dac/entity/model/rt"
	"dac/entity/utils/thttp"
	"encoding/json"
	"fmt"
	"strings"
	"trpc.group/trpc-go/trpc-go/log"
	cmdbPb "trpcprotocol/cmdb"
	collectorPb "trpcprotocol/collector"
)

// fetchGIDs 从本地缓存中查找已知的code-GID映射，返回已知映射和仍需远程获取的code列表
func (w *worker) fetchGIDs(codes []string) (rt.CodeGIDMapType, []string) {
	alreadyKnown := make(rt.CodeGIDMapType, len(codes))
	remainToFetch := make([]string, 0)

	w.RLock()
	defer w.RUnlock()

	for _, code := range codes {
		if gid, ok := w.codeGIDMap[code]; ok {
			alreadyKnown[code] = gid
		} else if len(code) > 0 {
			remainToFetch = append(remainToFetch, code)
		}
	}

	return alreadyKnown, remainToFetch
}

// FetchGID 获取单个设备code对应的GID
func (w *worker) FetchGID(code string) (db.GIDType, error) {
	if len(code) == 0 {
		return "", nil
	}

	codeGIDs, err := w.FetchGIDs([]string{code})
	if err != nil {
		return "", err
	}

	gid, ok := codeGIDs[code]
	if !ok {
		return "", fmt.Errorf("not found gid for code %s after fetch", code)
	}

	return gid, nil
}

// FetchGIDs 批量获取设备code对应的GID，优先从缓存获取，缓存未命中则从CMDB同步
func (w *worker) FetchGIDs(codes []string) (rt.CodeGIDMapType, error) {
	if !config.C.IsSyncGidFromCMDB() {
		return w.FetchGIDsOld(codes)
	}

	if len(codes) == 0 {
		return nil, nil
	}

	alreadyKnown, remainToFetch := w.fetchGIDs(codes)
	if len(remainToFetch) == 0 {
		return alreadyKnown, nil
	}

	// 从CMDB标准点配置获取GID
	fetchedGIDs, err := w.fetchGIDsFromCMDB(remainToFetch)
	if err != nil {
		return nil, fmt.Errorf("fetch gids from cmdb for codes %v error: %v", remainToFetch, err)
	}

	if len(remainToFetch) != len(fetchedGIDs) {
		config.Log.Warnf("request codes: %v, but return gids: %v", remainToFetch, fetchedGIDs)
	}
	if config.C.Debug {
		log.Infof("仍需要同步gid的设备code有：%v", remainToFetch)
	}

	w.setGIDs(fetchedGIDs)

	for code, gid := range fetchedGIDs {
		alreadyKnown[code] = gid
	}

	return alreadyKnown, nil
}

// fetchGIDsFromCMDB 从CMDB标准点配置获取GID
func (w *worker) fetchGIDsFromCMDB(codes []string) (rt.CodeGIDMapType, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	ctx := context.Background()
	result := make(rt.CodeGIDMapType)

	// codes包含门控器和门，需提取门控器code并去重
	controllerCodes := extractControllerCodes(codes)
	if len(controllerCodes) == 0 {
		return nil, fmt.Errorf("no valid controller codes extracted from %v", codes)
	}

	// 1. 调用CMDB接口获取标准点配置(获取门的gid)
	params, err := json.Marshal(controllerCodes)
	if err != nil {
		return nil, fmt.Errorf("marshal controller codes error: %v", err)
	}

	stdReq := &collectorPb.ReqFetchConfig{
		Params:    params,
		FetchType: collectorPb.ReqFetchConfig_FETCH_STD_POINTS,
	}

	configClient := collectorPb.NewConfigBusClientProxy()
	rsp, err := configClient.FetchConfig(ctx, stdReq)
	if err != nil {
		return nil, fmt.Errorf("fetch config from cmdb error: %v", err)
	}

	data := rsp.GetData()
	var configMap map[string]any
	err = json.Unmarshal(data, &configMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config data error: %v", err)
	}

	// 解析标准点配置，提取门的GID
	result, err = w.extractGIDFromConfig(configMap)
	if err != nil {
		return nil, fmt.Errorf("从configMap解析门gid映射失败，error: %v", err)
	}

	// 2. 调用CMDB接口获取采集设备配置(获取门控器的gid)
	controllerGIDs, err := w.fetchControllerGIDsFromCMDB(ctx, controllerCodes)
	if err != nil {
		log.Warnf("fetch controller gids from cmdb error: %v", err)
		// 获取门控器GID失败不影响门的GID同步，继续执行
	} else {
		// 合并门控器GID到结果中
		for code, gid := range controllerGIDs {
			result[code] = gid
		}
	}

	return result, nil
}

// fetchControllerGIDsFromCMDB 从CMDB采集设备配置获取门控器GID
func (w *worker) fetchControllerGIDsFromCMDB(ctx context.Context, controllerCodes []string) (rt.CodeGIDMapType, error) {
	if len(controllerCodes) == 0 {
		return nil, nil
	}

	result := make(rt.CodeGIDMapType)

	// 1. 获取需要从cmdb同步门禁的模组id列表
	mozuIDs := config.C.CMDBSyncMozus
	if len(mozuIDs) == 0 {
		return nil, fmt.Errorf("需要从cmdb同步门禁的模组id配置为空")
	}

	// 2. 构建需要查询的门控器code集合，用于过滤
	controllerCodeSet := make(map[string]bool)
	for _, code := range controllerCodes {
		controllerCodeSet[code] = true
	}

	// 3. 调用 ListCollectorDevice 接口获取门控器信息
	mozus := parseStringToInt32(mozuIDs)
	req := &cmdbPb.ReqListCollectorDevice{
		MozuId:        mozus,
		CollectorType: []int32{collectorDeviceTypeController}, // 5=门禁控制器类型
	}

	configClient := cmdbPb.NewConfigQueryClientProxy()
	rsp, err := configClient.ListCollectorDevice(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("list collector device from cmdb error: %v", err)
	}

	// 4. 解析响应，提取门控器code和gid的映射（只保留需要的门控器）
	for _, device := range rsp.GetList() {
		if device == nil {
			continue
		}
		deviceName := device.GetDeviceName()
		deviceGid := device.GetDeviceGid()
		// 只处理我们需要的门控器
		if deviceName != "" && deviceGid != "" && controllerCodeSet[deviceName] {
			result[deviceName] = db.GIDType(deviceGid)
		}
	}

	if config.C.Debug {
		log.Infof("此次从cmdb同步的门控器gidmap为：%v，一共有%v条", result, len(result))
	}

	return result, nil
}

// collectorDeviceTypeController 门禁控制器采集设备类型
const collectorDeviceTypeController = 5

// parseStringToInt32 将字符串切片转换为int32切片
func parseStringToInt32(strs []string) []int32 {
	result := make([]int32, 0, len(strs))
	for _, s := range strs {
		var num int
		if _, err := fmt.Sscanf(s, "%d", &num); err == nil {
			result = append(result, int32(num))
		}
	}
	return result
}

// extractControllerCodes 从codes中提取门控器code并去重
// 输入: ["controller1.GSM_1", "controller1.GSM_2", "controller2", "controller2.GSM_1"]
// 输出: ["controller1", "controller2"]
func extractControllerCodes(codes []string) []string {
	controllerMap := make(map[string]bool)
	for _, code := range codes {
		controllerCode := extractControllerCode(code)
		if controllerCode != "" {
			controllerMap[controllerCode] = true
		}
	}

	result := make([]string, 0, len(controllerMap))
	for code := range controllerMap {
		result = append(result, code)
	}
	return result
}

// extractControllerCode 提取门控器code
// 输入: "controller1.GSM_1" -> 输出: "controller1"
// 输入: "controller1" -> 输出: "controller1"
func extractControllerCode(code string) string {
	if idx := strings.Index(code, ".GSM_"); idx != -1 {
		return code[:idx]
	}
	return code
}

// extractGIDFromConfig 从标准点配置中提取采集GID
func (w *worker) extractGIDFromConfig(configMap map[string]any) (rt.CodeGIDMapType, error) {
	gidMap := make(rt.CodeGIDMapType)
	if len(configMap) == 0 {
		log.Warnf("std points config not exist")
		return gidMap, nil
	}
	// 解析配置数据
	var stdPoints rt.StdInstancePointsInfo
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

	err := json.Unmarshal(stdJson, &stdPoints)
	if err != nil {
		log.Error("std point config: unmarshal fail:%v", err)
		return nil, err
	}

	// 从所有的标准设备数据里解析gid
	// 每个门有四个测点，使用map去重：key为门采集code，value为门GID
	for _, stdPoint := range stdPoints {
		// 跳过无效数据
		if stdPoint.StdDevice == "" || stdPoint.MappingZh == "" || stdPoint.Mapping == "" {
			continue
		}

		// 从StdPointZh解析code，从Mapping解析gid
		doorCode := extractCodeOrGid(stdPoint.MappingZh)
		doorGID := extractCodeOrGid(stdPoint.Mapping)
		if doorCode == "" || doorGID == "" {
			log.Warnf("从标准点配置中解析门code和gid失败, StdPointZh=%s, Mapping=%s", stdPoint.StdPointZh, stdPoint.Mapping)
			continue
		}
		gidMap[doorCode] = db.GIDType(doorGID)
	}
	if config.C.Debug {
		log.Infof("此次从cmdb同步的门设备gidmap为：%v，一共有%v条", gidMap, len(gidMap))
	}

	return gidMap, nil
}

// extractCodeFromStdPointZh 从StdPointZh中提取门采集code
// 输入格式: "A=controller1.GSM_1.Comm" 或 "A=12345678.DoorState"
// 输出: "controller1.GSM_1" 或 "12345678"
func extractCodeOrGid(mapping string) string {
	if mapping == "" {
		return ""
	}

	// 按等号分割: "A=xxx.测点"
	parts := strings.Split(mapping, "=")
	if len(parts) != 2 {
		log.Warnf("不可用的mapping格式: %s, 应当为: A=xxx.测点", mapping)
		return ""
	}

	// 找到最后一个点号，提取之前的所有内容
	rightPart := strings.TrimSpace(parts[1])
	lastDotIndex := strings.LastIndex(rightPart, ".")
	if lastDotIndex == -1 {
		log.Warnf("不可用的mapping右侧格式: %s, 应当为: xxx.测点", mapping)
		return ""
	}

	// 提取最后一个点号之前的部分
	value := rightPart[:lastDotIndex]
	return value
}

// FetchGIDsOld 原从gidMapping服务同步gid（旧版接口，通过HTTP调用）
func (w *worker) FetchGIDsOld(codes []string) (rt.CodeGIDMapType, error) {
	if len(codes) == 0 {
		return nil, nil
	}

	alreadyKnown, remainToFetch := w.fetchGIDs(codes)
	if len(remainToFetch) == 0 {
		return alreadyKnown, nil
	}

	var (
		req struct {
			Codes []string `json:"collect_codes"`
		}
		resp rt.CodeGIDMapType
		err  error
	)
	req.Codes = remainToFetch

	if err = thttp.PostJSON(config.C.GIDMapping.URL.ConvertCode, req, 30000, &resp); err != nil {
		return nil, fmt.Errorf("get gid from codes %v error: %v", codes, err)
	}
	if len(req.Codes) != len(resp) {
		config.Log.Warnf("request codes: %v, but return gids: %v", req.Codes, resp)
	}

	w.setGIDs(resp)

	for code, gid := range resp {
		alreadyKnown[code] = gid
	}

	return alreadyKnown, nil
}
