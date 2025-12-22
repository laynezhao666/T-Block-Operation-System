package expressioncmd

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/cm"
	"agent/logic/collector/device/model"
	"agent/logic/collector/rtdb"
	rtdbModel "agent/logic/collector/rtdb/model"
	"agent/logic/plugin"
	"agent/utils"
	"agent/utils/osal"
	"common/util/expr"
	"strings"
	"sync"

	"trpc.group/trpc-go/trpc-go/log"
)

const (
	cmdExpression              string = "_expression"
	mappingListSplitChar       string = ","
	mappingExpressionSplitChar string = "="
)

var (
	once sync.Once
	plu  = expressionCmdPlugin{
		expressionCollectedPoints: make(model.StdInstancePointsInfo, 0),
	}
)

type expressionCmdPlugin struct {
	expressionCollectedPoints model.StdInstancePointsInfo
}

// Plugin 返回插件实例
func Plugin() *expressionCmdPlugin {
	return &plu
}

// Notify 通知插件事件
func (p *expressionCmdPlugin) Notify(event plugin.EventType) {
	if event == plugin.EventCollectConfigChange {
		p.refreshAllPoints()
	}
}

// Do 执行插件
func (p *expressionCmdPlugin) Do(arg any) {
	once.Do(
		func() {
			// 在启动插件时调用，此时配置一定已拉取完成
			p.refreshAllPoints()
		})
	ids := make(definition.DataPointIDsType, len(p.expressionCollectedPoints))
	for i, pointInfo := range p.expressionCollectedPoints {
		ids[i] = definition.DataPointIDType(pointInfo.StdPoint)
	}
	points := rtdb.GetDataPointsByID(ids)
	successCount := 0
	for i, pointInfo := range p.expressionCollectedPoints {
		val, qua, tms := getExpressionValue(pointInfo)
		if tms <= 0 {
			tms = utils.GetNowUTCTimeStamp()
		}
		points[i].Rtd.Val.Pv = osal.NewVariantWithValue(val)
		points[i].Rtd.Val.Qua = qua
		points[i].Rtd.Val.Tms = tms
		points[i].Rtd.Virtual = false
		points[i].ID = ids[i]
		points[i].DeviceGiD = definition.DeviceGidType(pointInfo.StdDevice)
		points[i].PointType = definition.CollectPointType
		if qua == consts.QualityOk {
			successCount = successCount + 1
		}
	}
	log.Debugf("expressionCmdPlugin: success/total[%v/%v] points: %+v", successCount, len(points), points)
	rtdb.SetDataPoints(points, true)
}

// ProcessRtd 处理rtd数据
func (p *expressionCmdPlugin) ProcessRtd(deviceID definition.DeviceGidType, points rtdbModel.DataPoints, ignore bool) {

}

func getExpressionValue(pointInfo model.StdInstancePointInfo) (any, consts.Quality, int64) {
	tms := int64(-1) // 采集时间
	// 变量映射
	parameters := make(map[string]any, len(pointInfo.Param))
	for k, v := range pointInfo.Param {
		var rtdVal rtdbModel.RTValue
		// 从rtdb获取
		pointId := []definition.DataPointIDType{definition.DataPointIDType(v)}
		dataPoints := rtdb.GetDataPointsByID(pointId)
		if len(dataPoints) != 1 {
			return nil, consts.QualityStdParamErr, tms
		}
		rtdVal = dataPoints[0].Rtd.Val
		if !rtdVal.IsOK() {
			return nil, rtdVal.Qua, tms
		}
		fv, err := rtdVal.Pv.AsFloat()
		if err != nil {
			return nil, consts.QualityValueTypeError, tms
		}
		parameters[k] = float64(fv)
		if rtdVal.Tms > tms {
			tms = rtdVal.Tms
		}
	}
	result, qua, err := expr.EvalStr(pointInfo.Expr, parameters)
	if err != nil {
		log.Debugf("plugin expr: pointInfo=%v, param=%v, err=%v", pointInfo, parameters, err)
	}
	return result, consts.Quality(qua), tms
}

// 重新加载需要计算的表达式采集点
func (p *expressionCmdPlugin) refreshAllPoints() {
	// 清空现有表达式测点列表
	p.expressionCollectedPoints = make([]model.StdInstancePointInfo, 0)

	templates := cm.Worker().CopyAllTemplateData()
	exprssionTemplatePoints := make(map[string]model.InstancePointsInfo)
	// key为测点中存在_expression的模板名，value为带表达式的采集点列表
	for key, t := range templates {
		l := make(model.InstancePointsInfo, 0)
		for _, p := range t.PointsInfo {
			if p.ProtocolDef.Command == cmdExpression {
				l = append(l, p)
			}
		}
		if len(l) != 0 {
			exprssionTemplatePoints[key] = l
		}
	}
	devices := cm.Worker().CopyAllDevices()
	// 构造成可以直接借用标准点计算函数的的带表达式的采集点
	for _, d := range devices {
		name := d.TemplateData.TemplateName
		if pointList, ok := exprssionTemplatePoints[name]; ok {
			for _, point := range pointList {
				id := string(definition.GenerateDataPointID(d.Gid, definition.PointIDType(point.ID)))
				params := getMappingParams(point.ExprDef.Mapping, string(d.Gid))
				p.expressionCollectedPoints = append(p.expressionCollectedPoints, model.StdInstancePointInfo{
					StdDevice: string(d.Gid),
					StdPoint:  id,
					Mapping:   point.ExprDef.Mapping,
					Expr:      point.ExprDef.Expr,
					Param:     params,
				})
			}

		}
	}
	log.Warnf("refresh expression collect point done: points count %v", len(p.expressionCollectedPoints))
}

// mapping字符串的样式"A=point1;B=point2;C=point3"
func getMappingParams(mappingStr string, pointId string) map[string]string {
	mpList := strings.Split(mappingStr, mappingListSplitChar)
	params := map[string]string{}
	for _, v := range mpList {
		pair := strings.Split(v, mappingExpressionSplitChar)
		if len(pair) != 2 {
			continue
		}
		// 对于设备gid为123的，此测点依赖的参数{"A":"123.point1",...}
		// 带有表达式计算的采集点，依赖的点都是同设备的其它点
		params[pair[0]] = string(definition.GenerateDataPointID(pointId, definition.PointIDType(pair[1])))
	}
	return params
}

// GetInterval 获取采集间隔
func (p *expressionCmdPlugin) GetInterval() int {
	return 3
}
