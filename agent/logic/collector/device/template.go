package device

import (
	"agent/entity/config"
	"agent/entity/consts"
	"encoding/json"
	"strings"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/definition"
	model3 "agent/entity/model"
	"agent/logic/collector/device/driver"
	"agent/logic/collector/device/model"
	model2 "agent/logic/collector/rtdb/model"
)

// MapCollectProtocolPacket 采集指令包
type MapCollectProtocolPacket map[string]*model.CollectProtocolPacket

// MapControlProtocolPacket 控制指令包
type MapControlProtocolPacket map[definition.DataPointIDType]*model.ControlProtocolPacket

// TemplateProtocol 模板协议
type TemplateProtocol struct {
	templateName string
	driverInfo   model.DriverInfo
	// 驱动
	driver driver.IDriver
	// 采集指令包
	collectPackets MapCollectProtocolPacket
	// 控制指令包
	controlPackets MapControlProtocolPacket
	// 采集指令包列表
	listPackets model.ListCollectPackets
	// 测点数量
	pointCount int
	// 记录模版里被引用来做表达式计算的测点,这里只支持单测点表达式
	expressionPoints map[string][]ExpressionInfo
}

// ExpressionInfo 表达式信息
type ExpressionInfo struct {
	Expr           string // 表达式
	KeyName        string // key名称
	DstPointNo     string // 目标测点名
	ValueType      string // 值类型
	ValuePrecision string // 值精度
}

// NewTemplateProtocol 根据 templateName 生成新的模板协议
func NewTemplateProtocol(templateName string) *TemplateProtocol {
	return &TemplateProtocol{
		templateName:     templateName,
		driverInfo:       model.DriverInfo{},
		collectPackets:   make(MapCollectProtocolPacket),
		controlPackets:   make(MapControlProtocolPacket),
		listPackets:      make(model.ListCollectPackets, 0, 10),
		pointCount:       0,
		expressionPoints: make(map[string][]ExpressionInfo),
	}
}

// Unload 卸载所有数据
func (t *TemplateProtocol) Unload() {
	if t == nil {
		return
	}
	t.collectPackets = nil
	t.controlPackets = nil
	t.listPackets = nil
	t.pointCount = 0
}

// GetPointCount 获取测点数量
func (t *TemplateProtocol) GetPointCount() int {
	return t.pointCount
}

// GetTemplateName 获取模板名称
func (t *TemplateProtocol) GetTemplateName() string {
	if t == nil {
		return ""
	}
	return t.templateName
}

// GetDrvInfo 获取驱动信息
func (t *TemplateProtocol) GetDrvInfo() model.DriverInfo {
	if t == nil {
		return model.DriverInfo{}
	}
	return t.driverInfo
}

// GetCollectPackets 获取采集指令包
func (t *TemplateProtocol) GetCollectPackets() model.ListCollectPackets {
	if t == nil {
		return nil
	}
	return t.listPackets
}

// GetDriver 获取驱动
func (t *TemplateProtocol) GetDriver() driver.IDriver {
	if t == nil {
		return nil
	}
	return t.driver
}

// Load 从 templateData 中加载模板协议
func (t *TemplateProtocol) Load(templateData *model3.TemplateData) bool {
	if t == nil || templateData == nil {
		return false
	}
	if !t.parseDriverInfo(templateData.DrvInfo) {
		return false
	}

	points := templateData.GetPoints()
	if !t.parsePointsInfo(points) {
		return false
	}
	t.pointCount = len(points)
	return true
}

// Reload 从 templateData 重新加载模板协议
func (t *TemplateProtocol) Reload(templateData *model3.TemplateData) bool {
	t.Unload()
	t.init()
	return t.Load(templateData)
}

func (t *TemplateProtocol) init() {
	t.collectPackets = make(MapCollectProtocolPacket)
	t.controlPackets = make(MapControlProtocolPacket)
	t.listPackets = make(model.ListCollectPackets, 0, 10)
}

// parseDriverInfo 解析驱动信息，解析成功则返回 true，否则返回 false
func (t *TemplateProtocol) parseDriverInfo(driverInfo model.DriverInfo) bool {
	if config.GetRB().IsSimulationEnable() {
		driverInfo.DriverName = consts.Simulator
	}

	t.driverInfo = driverInfo
	ok := false
	t.driver, ok = driver.DriverManager().GetDriver(strings.ToLower(driverInfo.DriverName))
	if !ok {
		log.Warnf("load driver: \"%v\" failed.", driverInfo.DriverName)
	}
	return ok
}

// parsePointsInfo 解析所有测点信息，解析成功则返回 true，否则返回 false
func (t *TemplateProtocol) parsePointsInfo(points model.InstancePointsInfo) bool {
	pointNoSet := make(map[string]struct{})
	calcPoints := make([]model.TemplateInstancePointInfo, 0)
	for _, point := range points {
		if !t.parsePointInfo(point) {
			return false
		}
		// 没有开直接计算，则跳过后面的逻辑
		if !config.GetRB().Task.Local.DirectCalc {
			continue
		}
		// 如果不是表达式计算测点则记录PointNo,否则记录下来做进一步处理
		if point.ProtocolDef.Command != model.CmdExpression {
			pointNoSet[point.ID.GetPointNo()] = struct{}{}
			continue
		}
		calcPoints = append(calcPoints, point)
	}
	for _, point := range calcPoints {
		ok, srcPointNo, keyName := point.ExprDef.MatchDirectCalc(pointNoSet)
		if ok {
			info := ExpressionInfo{
				Expr:           point.ExprDef.Expr,
				KeyName:        keyName,
				DstPointNo:     point.ID.GetPointNo(),
				ValuePrecision: point.ExprDef.Precision,
				ValueType:      point.ProtocolDef.Datatype,
			}
			// 添加到计算测点里
			if array, has := t.expressionPoints[srcPointNo]; has {
				t.expressionPoints[srcPointNo] = append(array, info)
			} else {
				t.expressionPoints[srcPointNo] = []ExpressionInfo{info}
			}
		}
	}
	return true
}

// parsePointInfo 解析测点信息，解析成功则返回 true，否则返回 false
func (t *TemplateProtocol) parsePointInfo(point model.TemplateInstancePointInfo) bool {
	var pointType model.PointType
	var valDesc interface{} = nil
	switch point.ValueType {
	case model.AnalogTypeString:
		pointType = model.AnalogType
		valDesc = t.createAnalogPointValDesc(&point)
	case model.DigitalTypeString:
		pointType = model.DigitalType
		valDesc = t.createDigitalPointValDesc(&point)
	case model.EnumTypeString:
		pointType = model.EnumType
		valDesc = t.createEnumPointValDesc(&point)
	case model.BoolTypeString:
		pointType = model.DigitalType
		valDesc = t.createDigitalPointValDesc(&point)
	default:
		log.Warnf("unknown value type: \"%v\", point: %+v", point.ValueType, point)
		return false
	}
	if valDesc == nil {
		log.Warnf("valDesc is nil")
		return false
	}

	// 表达式计算点由 expressionCmdPlugin 负责计算，不需要驱动解析器
	var valParser interface{}
	if point.ProtocolDef.Command == model.CmdExpression {
		log.Debugf("skip valParser for expression point: pointID=%s, expr=%s",
			point.ID, point.ExprDef.Expr)
	} else {
		valParser = t.createValParser(point)
		if valParser == nil {
			log.Warnf("valParser is nil: pointID=%s, point=%+v", point.ID, point)
			return false
		}
	}

	pointInfo := model.PointInfo{
		Attr: model.PointAttr{
			ID:        point.ID,
			Type:      pointType,
			ValDesc:   valDesc,
			ValParser: valParser,
		},
		RtVal: model2.NewRTValue(),
	}
	log.Debugf("load point: %+v, %+v", pointInfo, point.ProtocolDef)

	cmd := point.ProtocolDef.Command
	access := strings.ToUpper(point.Access)
	// 只有写属性的测点只加到控制点
	if model.AccessWrite == access {
		t.addControlPoint(cmd, &pointInfo)
		return true
	}
	t.addCollectPoint(cmd, &pointInfo)
	// 加入具有写权限的测点
	if strings.Index(access, model.AccessWrite) >= 0 {
		t.addControlPoint(cmd, &pointInfo)
	}
	return true
}

func (t *TemplateProtocol) createAnalogPointValDesc(point *model.TemplateInstancePointInfo) interface{} {
	desc := AnalogValueDesc{
		ScaleEnable: false,
	}
	if !desc.Parse(point) {
		return nil
	}
	return desc
}

func (t *TemplateProtocol) createDigitalPointValDesc(point *model.TemplateInstancePointInfo) interface{} {
	desc := make(DigitalValDesc)
	if !desc.Parse(point.ValueDef) {
		return nil
	}
	return desc
}

func (t *TemplateProtocol) createEnumPointValDesc(point *model.TemplateInstancePointInfo) interface{} {
	desc := make(EnumValDesc)
	if !desc.Parse(point.ValueDef) {
		return nil
	}
	return desc
}

// createValParser 创建值解析对象
func (t *TemplateProtocol) createValParser(point model.TemplateInstancePointInfo) interface{} {
	protocolDef := &point.ProtocolDef
	reg := protocolDef.Register

	simulatorRule := ""
	if config.GetRB().IsSimulationEnable() {
		switch temp := point.SimulatorDef.(type) {
		case string:
			simulatorRule = temp
		case []byte:
			simulatorRule = string(temp)
		default:
			tempBytes, err := json.Marshal(point.SimulatorDef)
			if err != nil {
				return nil
			}
			simulatorRule = string(tempBytes)
		}
		reg = simulatorRule
	}

	valParams := &model.ValParseParams{
		PointID:   string(point.ID),
		DataAddr:  reg,
		DataType:  protocolDef.Datatype,
		ByteOrder: protocolDef.Byteorder,
		Extend:    protocolDef.Extend,
	}
	return t.driver.CreateValParseObj(valParams)
}

// addCollectPoint 将所有使用 cmd 指令采集的 point 测点放入同一采集报文中
func (t *TemplateProtocol) addCollectPoint(cmd string, point *model.PointInfo) {
	packets, ok := t.collectPackets[cmd]
	if ok {
		packets.Points = append(packets.Points, point)
	} else {
		packet := &model.CollectProtocolPacket{
			Command: cmd,
			Points:  model.ListPoints{point},
		}
		t.collectPackets[cmd] = packet
		t.listPackets = append(t.listPackets, packet)
	}
}

// addControlPoint 将控制测点 point 放入采集报文
func (t *TemplateProtocol) addControlPoint(cmd string, point *model.PointInfo) {
	t.controlPackets[point.Attr.ID] = &model.ControlProtocolPacket{
		Command: cmd,
		Point:   point,
	}
}

func (t *TemplateProtocol) findCtlProtoPacket(pointGid definition.DataPointIDType) *model.ControlProtocolPacket {
	packet, ok := t.controlPackets[pointGid]
	if !ok {
		return nil
	}
	return packet
}
