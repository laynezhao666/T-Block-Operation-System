package alarm

import (
	"alarm-server/entity/constant"
	"alarm-server/logic/api/strategy"
	"alarm-server/repo/rpc"
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"etrpc-go/log"

	"github.com/samber/lo"

	cPb "trpcprotocol/alarm-compute"
	pb "trpcprotocol/alarm-server"
)

// VariableGidMap (告警/恢复)变量映射
type VariableGidMap struct {
	ExprMap map[string]string `json:"expr_map"`
	Engine  string            `json:"engine"`
}

// ExpressionsMap (告警/恢复)表达式映射
type ExpressionsMap struct {
	Fire    VariableGidMap `json:"fire"`
	Restore VariableGidMap `json:"restore"`
}

func (a *alarmLogicImpl) AlarmDiagnose(ctx context.Context, req *pb.ReqAlarmDiagnose) (*pb.RspAlarmDiagnose, error) {
	strategyReq := &pb.ReqStrategyInstance{
		MozuId:       req.MozuId,
		Rid:          req.Rid,
		DeviceGid:    req.DeviceGid,
		DeviceNumber: req.DeviceNumber,
	}
	strategyList, err := strategy.NewStrategyLogicApi().GetStrategyInstance(ctx, strategyReq)
	if err != nil {
		return nil, fmt.Errorf("AlarmDiagnose 读取策略列表失败, err: %s", err.Error())
	}
	if len(strategyList.List) == 0 {
		return nil, fmt.Errorf("AlarmDiagnose 读取策略列表为空")
	}
	alarmExpEn, alarmExpZh := strategyList.List[0].AlarmExpStd, strategyList.List[0].AlarmExp
	if len(alarmExpEn) == 0 || len(alarmExpZh) == 0 {
		return nil, fmt.Errorf("AlarmDiagnose 策略告警表达式无效")
	}
	fExprItemEnZhMap := map[string]string{}
	// 获取表达式 英文 -> 中文 映射
	alarmsEn, alarmsZh := a.splitExpr(alarmExpEn), a.splitExpr(alarmExpZh)
	if len(alarmsEn) != len(alarmsZh) {
		return nil, fmt.Errorf("AlarmDiagnose 中英文策略表达式解析不一致，英文: %v, 中文: %v", alarmExpEn, alarmExpZh)
	}
	for i := range alarmsEn {
		fExprItemEnZhMap[alarmsEn[i]] = alarmsZh[i]
	}
	// 调用接口，进行告警回放
	reqComExpressList := [][]string{}
	reqComPMapList := []map[string]string{}
	for _, strategy := range strategyList.List {
		reqComExpressList = append(reqComExpressList, alarmsEn)
		reqComPMapList = append(reqComPMapList, a.parsePMap(strategy.ExpressionMap))
	}
	reqComPv := map[string]map[int64]float64{}
	for k, v := range req.Pv {
		reqComPv[k] = v.Tv
	}
	reqComInterval, err := a.geneInterval(req.StartTime, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("AlarmDiagnose 生成计算间隔失败, err: %s", err.Error())
	}
	comRsp, err := rpc.NewAlarmComputeRpc().ExpCompute(ctx, req.StartTime, req.EndTime, int32(reqComInterval),
		reqComExpressList, reqComPMapList, reqComPv)
	if err != nil {
		return nil, fmt.Errorf("AlarmDiagnose 调用表达式计算服务失败, err: %s", err.Error())
	}
	if len(comRsp.List) != len(alarmsEn)*len(strategyList.List) {
		return nil, fmt.Errorf("AlarmDiagnose 响应结果与请求表达式数量不一致，req: %v, rsp: %v", len(reqComExpressList),
			len(comRsp.List))
	}
	// 将计算结果按策略维度拆分
	strategyComRspList := lo.Chunk(comRsp.List, len(alarmsEn))
	rsp := &pb.RspAlarmDiagnose{
		MozuId: req.MozuId,
		Rid:    req.Rid,
		List:   []*pb.RspAlarmDiagnose_Item{},
	}
	for index, strategy := range strategyList.List {
		comRspItem := strategyComRspList[index]
		for _, comItem := range comRspItem {
			rspItem := &pb.RspAlarmDiagnose_Item{
				DeviceGid:    strategy.DeviceGid,
				DeviceNumber: strategy.DeviceNumber,
				ExpressStd:   comItem.Express,
				ExpressStr:   fExprItemEnZhMap[comItem.Express],
				CalRes: lo.Map(comItem.CalRes, func(item *cPb.RspExpCompute_Item_StepResult, index int) *pb.RspAlarmDiagnose_Item_StepResult {
					return &pb.RspAlarmDiagnose_Item_StepResult{
						Timestamp: item.Timestamp,
						Success:   item.Success,
						Fired:     item.Fired,
					}
				}),
			}
			rsp.List = append(rsp.List, rspItem)
		}
	}
	return rsp, nil
}

// splitExpr 拆分表达式
func (a *alarmLogicImpl) splitExpr(expr string) []string {
	// 最终结果包含 总表达式 子表达式
	res := []string{expr}
	reg, _ := regexp.Compile(`&&|\|\|`)
	exprItems := reg.Split(expr, -1)
	if len(exprItems) == 1 {
		return res
	}
	// 需要采取补括号的方式，因为会遇到类似【!(FanAlarm==1)】的表达式
	for i := range exprItems {
		leftCount, rightCount := 0, 0
		for _, char := range exprItems[i] {
			if char == '(' {
				leftCount++
			}
			if char == ')' {
				rightCount++
			}
		}
		diff := leftCount - rightCount
		if diff == 0 {
			continue
		}
		brackets := ""
		if diff > 0 {
			// 左边的括号多了
			for i := 0; i < diff; i++ {
				brackets = brackets + ")"
			}
			exprItems[i] = exprItems[i] + brackets
		} else {
			for i := 0; i < -diff; i++ {
				brackets = "(" + brackets
			}
			exprItems[i] = brackets + exprItems[i]
		}
		itemLen := len(exprItems[i])
		for l, r := 0, itemLen-1; l <= r; {
			if l < itemLen-1 && exprItems[i][l] == byte('!') && exprItems[i][l+1] == byte('(') {
				// 不处理取反
				l++
			} else if exprItems[i][l] == byte('(') && exprItems[i][r] == byte(')') {
				l, r = l+1, r-1
			} else {
				exprItems[i] = string(exprItems[i][l : r+1])
				break
			}
		}
	}
	res = append(res, exprItems...)
	return res
}

// 从expression_map 读取告警pMap
func (a *alarmLogicImpl) parsePMap(expressionMap string) map[string]string {
	// 将expressionMap 解析为 ExpressionsMap 结构体
	var expMap *ExpressionsMap
	if err := json.Unmarshal([]byte(expressionMap), &expMap); err != nil {
		log.Errorf("parsePMap:%s, err: %v", expressionMap, err)
		return map[string]string{}
	}
	return expMap.Fire.ExprMap
}

// geneInterval 根据开始和结束时间，生成interval
func (a *alarmLogicImpl) geneInterval(beginTime, endTime int64) (interval int, err error) {
	duration := endTime - beginTime
	if duration < 0 {
		err = fmt.Errorf("bad duration: %v", duration)
		return
	}
	durationInt := int(duration)
	if durationInt > constant.SecondsInDay {
		// 超过一天，10min计算一次
		interval = 10 * constant.MinuteInterval
		return
	}
	// 超过一小时，5min计算一次
	if durationInt > 60*constant.SecondsInMinute {
		interval = 5 * constant.MinuteInterval
		return
	}
	if durationInt > 10*constant.SecondsInMinute {
		// 超过10分钟，10秒计算一次
		interval = constant.SecondInterval
		return
	}
	interval = constant.DefaultInterval
	return
}
