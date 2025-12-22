package device

import (
	"encoding/json"
	"agent/entity/config"
	"agent/entity/consts"
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
}

// NewTemplateProtocol 根据 templateName 生成新的模板协议
func NewTemplateProtocol(templateName string) *TemplateProtocol {
	return &TemplateProtocol{
		templateName:   templateName,
		driverInfo:     model.DriverInfo{},
		collectPackets: make(MapCollectProtocolPacket),
		controlPackets: make(MapControlProtocolPacket),
		listPackets:    make(model.ListCollectPackets, 0, 10),
		pointCount:     0,
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
	for _, point := range points {
		if !t.parsePointInfo(point) {
			return false
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

	valParser := t.createValParser(point)
	if valParser == nil {
		return false
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

	cmd := point.ProtocolDef.Command
	// 默认所有测点均具有读权限
	t.addCollectPoint(cmd, &pointInfo)
	// 加入具有写权限的测点
	if rw := strings.ToUpper(point.Access); strings.Index(rw, model.AccessWrite) >= 0 {
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
