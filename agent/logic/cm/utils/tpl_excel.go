package utils

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"agent/entity/config"
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/entity/errcode"
	"agent/entity/model"
	cm2 "agent/logic/cm"
	cmodel "agent/logic/collector/device/model"
	tbosIo "agent/utils/file/io"

	"github.com/tealeg/xlsx/v3"
	"github.com/xuri/excelize/v2"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	thttp "trpc.group/trpc-go/trpc-go/http"
	"trpc.group/trpc-go/trpc-go/log"
)

const (
	DriverTypeModbus        = "MODBUS"
	V2SheetVersionName      = "版本"
	V2SheetDeviceInfoName   = "设备信息"
	V2SheetPointsName       = "通讯点表"
	V2SheetExpressionName   = "表达式计算"
	TemplateClassTitle      = "设备类型"
	TemplateVendorTitle     = "设备厂商"
	TemplateDeviceCodeTitle = "设备型号"
	TemplateDriverTitle     = "通讯协议名称"
	TemplateVersionTitle    = "协议版本"
	CommandAndFuncCodeField = "cmd"
	RegAddrField            = "reg"
	DataTypeField           = "datatype"
	ByteOrderField          = "byteorder"
	ExtendedArgField        = "ext"
	WeightField             = "weight"
)

const (
	// 通讯点表起始行号
	pointValueIndexRow = 2
	// 标题行号
	deviceTitleIndexRow = 0
	// 值行号
	driverValueIndexRow = 1
	// 设备类型列号
	typeIndexCol = 0
	// 厂商列号
	vendorIndexCol = 1
	// 协议名称列号
	driverLibIndexCol = 3
	// 协议版本列号
	protocolVersionIndexCol = 4
	// 扩展参数列号
	extendIndexCol = 5
	// 表达式计算sheet起始行号
	expressionIndexRow = 2
)

const (
	simulationRuleRegExp       = `^([a-zA-Z]+)\((.*)\)$`
	simulationFunctionRandom   = "random"
	simulationFunctionMonotone = "monotone"
	simulationFunctionStatic   = "static"
	paramsSeperator            = ","
	randomTypeUniform          = "uniform"
)

// SimulationRule 规则
type SimulationRule struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
	Max   string `json:"max,omitempty"`
	Min   string `json:"min,omitempty"`
	Step  string `json:"step,omitempty"`
	Type  string `json:"type,omitempty"`
}

func parseSimulationRule(simulationRuleStr string) *SimulationRule {
	if simulationRuleStr == "" {
		// 表达式点没有规则，返回默认值指针，否则json文件中的simulator字段是null
		return &SimulationRule{}
	}
	rule := &SimulationRule{}
	re := regexp.MustCompile(simulationRuleRegExp)
	matches := re.FindStringSubmatch(simulationRuleStr)

	// matches[0] 是完整匹配的字符串
	// matches[1] 是第一个括号捕获的内容，即函数名
	// matches[2] 是第二个括号捕获的内容，即参数列表
	if matches == nil || len(matches) != 3 {
		log.Errorf("parse simulation rule <%v> failed", simulationRuleStr)
		return nil
	}

	rule.Name = matches[1]
	switch matches[1] {
	// random(uniform,235,283)
	case simulationFunctionRandom:
		params := strings.Split(matches[2], paramsSeperator)
		// 暂时不支持非uniform的随机函数
		if params[0] != randomTypeUniform || len(params) != 3 {
			log.Errorf("parse simulation rule <%v> failed", simulationRuleStr)
			return nil
		}
		rule.Value = params[0]
		rule.Min = params[1]
		rule.Max = params[2]
	// monotone(0,1000000,0.1)
	case simulationFunctionMonotone:
		params := strings.Split(matches[2], paramsSeperator)
		if len(params) != 3 {
			log.Errorf("parse simulation rule <%v> failed", simulationRuleStr)
			return nil
		}
		rule.Min = params[0]
		rule.Max = params[1]
		rule.Step = params[2]
	// static(135.0)
	case simulationFunctionStatic:
		rule.Value = matches[2]
	default:
		log.Errorf("not supported simulation rule <%v>", simulationRuleStr)
	}
	return rule
}

// ProcessImportTemplate 处理导入的驱动模版文件
func ProcessImportTemplate(filesHeaders []*multipart.FileHeader) error {
	if config.GetRB().Project.Source != "local" {
		return errs.New(errcode.ErrBadRequest, "当前不是local模式，无法执行该操作")
	}
	templatesMap := cm2.Worker().CopyAllTemplateData()
	// 解析文件内容，转换为map[string]*model.TemplateData
	for _, fileHeader := range filesHeaders {
		fileName := fileHeader.Filename
		// 将文件名作为模版名称
		templateName := fileName[:len(fileName)-len(filepath.Ext(fileName))]
		file, err := fileHeader.Open()
		if err != nil {
			return errs.New(errcode.ErrCgiTemplateFileFail, fmt.Sprintf(
				"模版名称:%s,错误:%s", templateName, err.Error()))
		}
		// 表格解析
		excelPoints, driverInfo, err := ProcessTemplate(file, fileHeader)
		if err != nil {
			return errs.New(errcode.ErrCgiTemplateFileFail, fmt.Sprintf(
				"模版名称:%s,错误:%s", templateName, err.Error()))
		}
		pointsInfo := make(cmodel.InstancePointsInfo, 0, len(excelPoints))
		for _, ePoint := range excelPoints {
			// 值定义格式转换
			valueDef, err := StringToMap(ePoint.CollectTemplatePointModel)
			if err != nil {
				return errs.New(errcode.ErrCgiTemplateFileFail, err.Error())
			}
			// 读写属性转换
			rw, err := RwTrans(ePoint.ValueRW)
			if err != nil {
				return errs.New(errcode.ErrCgiTemplateFileFail, err.Error())
			}
			// 类型转换
			vt := ConvertDataType(ePoint.ValueType)
			simulationRule := parseSimulationRule(ePoint.SimulationRule)
			pointsInfo = append(pointsInfo, cmodel.TemplateInstancePointInfo{
				ID:     definition.DataPointIDType(ePoint.SignIdentifier),
				Access: rw,
				Name:   ePoint.SignName,
				// 表达式定义
				ExprDef: cmodel.ExpressionDefinition{
					Expr:    ePoint.Expression,
					Mapping: ePoint.ValueMap,
				},
				// 协议定义
				ProtocolDef: cmodel.ProtocolDefinition{
					Byteorder: ePoint.ByteOrder,
					// 采集指令，依赖于具体驱动的解释
					Command:  ePoint.Cmd,
					Datatype: ePoint.DataType,
					// 扩展参数
					Extend:            ePoint.ExtendFunc,
					Register:          ePoint.Address,
					Offset:            ePoint.Offset,
					Scale:             ePoint.Scale,
					CmdIntervalWeight: ePoint.CmdSendIntervalWeight,
				},
				ValueDef:      valueDef,
				ValueType:     vt,
				SimulatorDef:  simulationRule,
				SubDevice:     "",
				IsNorthDef:    ePoint.IsNorthDefinition,
				ValueRange:    ePoint.ValueRange,
				ValueDeadZone: ePoint.ValueDeadZone,
			})
		}
		tpl := &model.TemplateData{
			DrvInfo:    *driverInfo,
			PointsInfo: pointsInfo,
		}
		// 用模版名称作为key进行覆盖
		templatesMap[templateName] = tpl
	}
	// 保存到本地文件
	err := cm2.Worker().SaveTemplatesConfig(templatesMap)
	if err != nil {
		return errs.New(errcode.ErrCgiTemplateFileFail, err.Error())
	}
	cm2.NotifyConfigChange()
	return nil
}

// ProcessTemplate 处理单个驱动模版文件
func ProcessTemplate(file multipart.File, fileHeader *multipart.FileHeader) ([]CommonPointModel,
	*cmodel.DriverInfo, error) {

	xlsxFile, err := xlsx.OpenReaderAt(file, fileHeader.Size)
	if err != nil {
		return nil, nil, err
	}
	if len(xlsxFile.Sheets) == 0 {
		//return fmt.Errorf("%v: %v", reqParseFileOpenFailMsg, "工作表为空")
	}
	// 解析表格内容
	driverPointModels, driverInfo, err := ProcessTemplateXlsx(xlsxFile, fileHeader)

	if err != nil {
		return nil, nil, err
	}
	return driverPointModels, driverInfo, nil
}

// ProcessTemplateXlsx 处理单个驱动模版文件
func ProcessTemplateXlsx(xlsxFile *xlsx.File, fileHeader *multipart.FileHeader) ([]CommonPointModel, *cmodel.DriverInfo, error) {
	return ImportCollectTemplate(&ExcelTemplatePacket{XlsxFile: xlsxFile, FileName: fileHeader.Filename})
}

// ExcelTemplatePacket 采集模版表格结构
type ExcelTemplatePacket struct {
	XlsxFile *xlsx.File // excel文件
	FileName string     // excel文件名
	Oid      int

	TemplateName  string // excel模版名
	TemplateID    int    // Db模版Id
	Driver        string // 驱动名
	DriverVersion string // 驱动版本号
	//CollectDeviceTypeMap map[string]model.DeviceTypeModel

	ExcelPoints []CollectTemplatePointModel // excel行数据
	//DbPoints    []model.CollectorTemplatePoint       // db行数据
}

// CollectTemplatePointModel 驱动测点模型
type CollectTemplatePointModel struct {
	SubDevice             string `json:"sub_device" xlsx:"0"`
	SignIdentifier        string `json:"sign_identifier" xlsx:"1"`
	SignName              string `json:"sign_name" xlsx:"2"`
	ValueType             string `json:"value_type" xlsx:"3"`
	ValueRW               string `json:"value_rw" xlsx:"4"`
	ValueUnit             string `json:"value_unit" xlsx:"5"`
	ValueDesc             string `json:"value_desc" xlsx:"6"`
	ValueRange            string `json:"value_range" xlsx:"7"`
	ValueDeadZone         string `json:"value_dead_zone" xlsx:"8"`
	Cmd                   string `json:"cmd" xlsx:"9"`
	Address               string `json:"address" xlsx:"10"`
	DataType              string `json:"data_type" xlsx:"11"`
	ByteOrder             string `json:"byte_order" xlsx:"12"`
	Scale                 string `json:"scale" xlsx:"13"`
	Offset                string `json:"offset" xlsx:"14"`
	ExtendFunc            string `json:"extend_func" xlsx:"15"`
	CmdSendIntervalWeight string `json:"cmd_send_interval_weight" xlsx:"16"`
	IsNorthDefinition     string `json:"is_north_definition" xlsx:"17"`
	// Remark                string `json:"remark" xlsx:"18"`
	SimulationRule string `json:"simulation_rule" xlsx:"18"`
}

// ExpressionPointModel "表达式计算"sheet的测点
type ExpressionPointModel struct {
	SubDevice         string `json:"sub_device" xlsx:"0"`
	SignIdentifier    string `json:"sign_identifier" xlsx:"1"`
	SignName          string `json:"sign_name" xlsx:"2"`
	ValueType         string `json:"value_type" xlsx:"3"`
	ValueRW           string `json:"value_rw" xlsx:"4"`
	ValueUnit         string `json:"value_unit" xlsx:"5"`
	ValueDesc         string `json:"value_desc" xlsx:"6"`
	ValueRange        string `json:"value_range" xlsx:"7"`
	ValueDeadZone     string `json:"value_dead_zone" xlsx:"8"`
	Expression        string `json:"expression" xlsx:"9"`
	ValueMap          string `json:"value_map" xlsx:"10"`
	IsNorthDefinition string `json:"is_north_definition" xlsx:"11"`
}

// CommonPointModel 通用测点，"通讯点表"sheet的测点或"表达式计算"sheet的测点
type CommonPointModel struct {
	CollectTemplatePointModel
	Expression string `json:"expression"`
	ValueMap   string `json:"value_map"`
}

// ImportCollectTemplate 导入统一驱动模板处理逻辑
func ImportCollectTemplate(excelTemplatePacket *ExcelTemplatePacket) ([]CommonPointModel, *cmodel.DriverInfo, error) {
	excelTemplatePacket.FileName = filepath.Base(excelTemplatePacket.FileName)
	// 驱动信息
	driverInfo, err := getDriverSheetPoints(excelTemplatePacket.XlsxFile.Sheet[V2SheetDeviceInfoName])
	if err != nil {
		return nil, nil, err
	}

	// 测点信息
	excelPoints, err := GetTemplateSheetPoints(driverInfo.DriverName,
		excelTemplatePacket.XlsxFile.Sheet[V2SheetPointsName])
	if err != nil {
		return nil, nil, err
	}

	// 表达式
	expressionPoints, err := getExpressionSheetPoints(excelTemplatePacket.XlsxFile.Sheet[V2SheetExpressionName])
	if err != nil {
		return nil, nil, err
	}

	// 合并excelPoints和expressPoints
	mergedPoints, err := mergeExcelPointsAndExpressionPoints(excelPoints, expressionPoints)
	if err != nil {
		return nil, nil, err
	}
	return mergedPoints, driverInfo, nil
}

// GetTemplateSheetPoints 新模板sheet通用解析方法
func GetTemplateSheetPoints(protocolName string, sheet *xlsx.Sheet) ([]CollectTemplatePointModel, error) {
	if sheet == nil {
		return nil, fmt.Errorf("get template sheet error")
	}
	var point CollectTemplatePointModel
	excelPoints := make([]CollectTemplatePointModel, 0, sheet.MaxRow)
	for i := pointValueIndexRow; i < sheet.MaxRow; i++ {
		row, err := sheet.Row(i)
		if err != nil {
			return nil, err
		}
		if err = row.ReadStruct(&point); err != nil {
			return nil, err
		}
		if point.SignIdentifier == "" {
			continue // 可能会读到excel中的空行，需要跳过
		}
		excelPoints = append(excelPoints, point)
	}

	return parsePoints(protocolName, excelPoints)
}

// getExpressionSheetPoints 表达式sheet通用解析方法
func getExpressionSheetPoints(sheet *xlsx.Sheet) ([]ExpressionPointModel, error) {
	if sheet == nil {
		return nil, fmt.Errorf("get template sheet error")
	}
	var point ExpressionPointModel
	expressionList := []ExpressionPointModel{}
	for i := expressionIndexRow; i < sheet.MaxRow; i++ {
		row, err := sheet.Row(i)
		if err != nil {
			return nil, err
		}
		if err = row.ReadStruct(&point); err != nil {
			return nil, err
		}
		if point.SignIdentifier == "" {
			continue // 可能会读到excel中的空行，需要跳过
		}
		expressionList = append(expressionList, point)
	}
	return expressionList, nil
}

// mergeExcelPointsAndExpressionPoints 合并excelPoints和expressionPoints
func mergeExcelPointsAndExpressionPoints(collectPoints []CollectTemplatePointModel, expressionPoints []ExpressionPointModel) ([]CommonPointModel, error) {
	mergedPoints := []CommonPointModel{}
	pointMap := map[string]bool{}
	for _, point := range collectPoints {
		if pointMap[point.SignIdentifier] {
			return nil, fmt.Errorf("信号标识符重复: %+v", point.SignIdentifier)
		}
		// 采集点，直接赋值嵌套的字段
		mergedPoints = append(mergedPoints, CommonPointModel{
			CollectTemplatePointModel: point,
			Expression:                "",
			ValueMap:                  "",
		})
		pointMap[point.SignIdentifier] = true
	}
	for _, point := range expressionPoints {
		if pointMap[point.SignIdentifier] {
			return nil, fmt.Errorf("信号标识符重复: %+v", point.SignIdentifier)
		}
		// 表达点，逐个赋值字段
		mergedPoints = append(mergedPoints, CommonPointModel{
			CollectTemplatePointModel: CollectTemplatePointModel{
				SubDevice:         point.SubDevice,
				SignIdentifier:    point.SignIdentifier,
				SignName:          point.SignName,
				ValueType:         point.ValueType,
				ValueRW:           point.ValueRW,
				ValueUnit:         point.ValueUnit,
				ValueDesc:         point.ValueDesc,
				ValueRange:        point.ValueRange,
				ValueDeadZone:     point.ValueDeadZone,
				IsNorthDefinition: point.IsNorthDefinition,
			},
			Expression: point.Expression,
			ValueMap:   point.ValueMap,
		})
	}
	return mergedPoints, nil
}

type pointKey struct {
	SubDevice      string
	SignIdentifier string
}

func parsePoints(protocolName string, points []CollectTemplatePointModel) ([]CollectTemplatePointModel, error) {
	var exist bool
	var setMember struct{}
	lenPoints := len(points)
	resultPoints := make([]CollectTemplatePointModel, 0, lenPoints)
	pointsSet := make(map[pointKey]struct{}, lenPoints)
	for i := 0; i < lenPoints; i++ {
		point := points[i]

		if isIgnoreV2Point(&point) {
			continue
		}

		key := makePointKey(&point)
		if _, exist = pointsSet[key]; exist {
			return nil, fmt.Errorf("point repeated, key = %+v", key)
		}

		resultPoints = append(resultPoints, point)
		pointsSet[key] = setMember
	}

	err := driverPointsProcess(protocolName, resultPoints)
	if err != nil {
		return nil, err
	}
	return resultPoints, nil
}

// isIgnoreV2Point 判断是否忽略该测点
func isIgnoreV2Point(templatePoint *CollectTemplatePointModel) bool {
	if templatePoint == nil {
		return true
	}
	return len(templatePoint.Cmd) == 0 &&
		len(templatePoint.Address) == 0
}

func makePointKey(p *CollectTemplatePointModel) pointKey {
	return pointKey{SubDevice: p.SubDevice, SignIdentifier: p.SignIdentifier}
}

// driverPointsProcess 不同驱动特殊处理，暂时只有modbus需要处理
func driverPointsProcess(driverName string, points []CollectTemplatePointModel) error {
	switch strings.ToUpper(driverName) {
	case DriverTypeModbus:
		err := processModbusPoint(points)
		if err != nil {
			return err
		}
	}

	return nil
}

func processModbusPoint(points []CollectTemplatePointModel) error {
	pointList := make([]MeasurePoint, 0, len(points))
	for i := range points {
		pointMap := make(MapObject, 0)
		pointMap[CommandAndFuncCodeField] = points[i].Cmd
		pointMap[DataTypeField] = points[i].DataType
		pointMap[RegAddrField] = points[i].Address

		mp := MeasurePoint{ProtocolDescKey: pointMap}
		pointList = append(pointList, mp)
	}

	err := GenerateCommands(pointList)
	if err != nil {
		return err
	}

	// 重新设置cmd
	for i := range points {
		proto, ok := pointList[i][ProtocolDescKey].(MapObject)
		if !ok {
			return fmt.Errorf("point %d cmd protocol error, cmd = %s, addr = %s", i, points[i].Cmd, points[i].Address)
		}

		cmd, ok := proto[CommandAndFuncCodeField].(string)
		if !ok {
			return fmt.Errorf("point %d cmd error, cmd = %s, addr = %s", i, points[i].Cmd, points[i].Address)
		}
		points[i].Cmd = cmd
	}
	return nil
}

func getDriverSheetPoints(sheet *xlsx.Sheet) (*cmodel.DriverInfo, error) {
	if sheet == nil {
		return nil, fmt.Errorf("get driver sheet error")
	}
	// 类型
	deviceType, err := getValue(sheet, TemplateClassTitle, deviceTitleIndexRow, typeIndexCol, driverValueIndexRow,
		typeIndexCol)
	if err != nil {
		return nil, fmt.Errorf("get deviceType error: %w", err)
	}
	// 厂商
	vendor, err := getValue(sheet, TemplateVendorTitle, deviceTitleIndexRow, vendorIndexCol, driverValueIndexRow,
		vendorIndexCol)
	if err != nil {
		return nil, fmt.Errorf("get vendor error: %w", err)
	}
	// 协议名称
	driver, err := getValue(sheet, TemplateDriverTitle, deviceTitleIndexRow, driverLibIndexCol, driverValueIndexRow,
		driverLibIndexCol)
	if err != nil {
		return nil, fmt.Errorf("get driver error: %w", err)
	}
	// 协议版本
	version, err := getValue(sheet, TemplateVersionTitle, deviceTitleIndexRow, protocolVersionIndexCol,
		driverValueIndexRow, protocolVersionIndexCol)
	if err != nil {
		return nil, fmt.Errorf("get version error: %w", err)
	}
	// 扩展参数
	extend, err := getValue(sheet, TemplateVersionTitle, deviceTitleIndexRow, extendIndexCol,
		driverValueIndexRow, extendIndexCol)
	if err != nil {
		return nil, fmt.Errorf("get extend error: %w", err)
	}
	// 封装
	driverInfo := &cmodel.DriverInfo{
		Class:           deviceType,
		Vendor:          vendor,
		DriverName:      driver,
		ProtocolVersion: version,
		Extend:          extend,
	}
	return driverInfo, nil
}

func getValue(s *xlsx.Sheet, titleName string, titleIndexRow, titleIndexCol, valueIndexRow,
	valueIndexCol int) (string, error) {
	title, err := s.Cell(titleIndexRow, titleIndexCol)
	if err != nil {
		return "", fmt.Errorf("get %v error: %w", titleName, err)
	}
	if title.String() != titleName {
		return "", nil
	}

	value, err := s.Cell(valueIndexRow, valueIndexCol)
	if err != nil {
		return "", err
	}

	v, err := value.FormattedValue()
	if err != nil {
		return "", err
	}
	return v, nil
}

// ExportAllTemplates 导出所有模板
func ExportAllTemplates(ctx context.Context) (err error) {
	// 1. 遍历所有文件，导出到一个文件夹下并压缩
	templateList, err := os.ReadDir(filepath.Join(config.GetRB().GetProjectLocalPath(), consts.RelativeTemplateDir))
	if err != nil {
		return errs.New(errcode.ErrServerLogic, fmt.Sprintf("读取模板文件夹错误:%s", err.Error()))
	}
	tempDir, err := os.MkdirTemp("", "template-*")
	defer os.RemoveAll(tempDir)
	if err != nil {
		return errs.New(errcode.ErrServerLogic, fmt.Sprintf("创建临时文件夹错误:%s", err.Error()))
	}
	// 遍历模板文件并导出，压缩为zip再导出
	for _, tpl := range templateList {
		if !strings.HasSuffix(tpl.Name(), ".json") {
			continue
		}
		tplName := strings.TrimSuffix(tpl.Name(), ".json")
		path, err := ExportTemplate(ctx, tplName)
		if err != nil {
			log.Errorf("导出模板文件 %s 错误: %s", tpl.Name(), err.Error())
			continue
		}
		if err = os.Rename(path, filepath.Join(tempDir, tplName+".xlsx")); err != nil {
			log.Errorf("移动导出后的模板文件 %s 错误: %s", path, err.Error())
			continue
		}
	}
	tmpZipFilePath := filepath.Join(os.TempDir(), "templates.zip")
	defer os.Remove(tmpZipFilePath)
	err = zipFolder(tempDir, tmpZipFilePath)
	if err != nil {
		return errs.New(errcode.ErrServerLogic, fmt.Sprintf("压缩模板文件错误:%s", err.Error()))
	}

	// 2. 返回zip文件内容
	f, err := os.Open(tmpZipFilePath)
	if err != nil {
		return errs.New(errcode.ErrServerLogic, fmt.Sprintf("读取压缩文件错误:%s", err.Error()))
	}
	msg := trpc.Message(ctx)
	msg.WithSerializationType(codec.SerializationTypeUnsupported)
	head := thttp.Head(ctx)
	head.Response.Header().Set("Content-Type", "application/zip")
	head.Response.Header().Set("Content-Disposition", "attachment; filename="+"templates.zip")
	_, err = io.Copy(head.Response, f)
	if err != nil {
		return errs.New(errcode.ErrServerLogic, fmt.Sprintf("拷贝压缩文件错误:%s", err.Error()))
	}
	return nil
}

// ExportTemplate 下载一个excel文件
func ExportTemplate(ctx context.Context, fileName string) (excelPath string, err error) {
	// 1. 读取json文件和excel模板
	jsonPath := filepath.Join(config.GetRB().GetProjectLocalPath(), consts.RelativeTemplateDir, fileName+".json")
	var jsonContent model.TemplateData
	err = tbosIo.JSON.Read(jsonPath, &jsonContent)
	if err != nil {
		return "", errs.New(errcode.ErrServerLogic, fmt.Sprintf("读取json文件错误: %s", err.Error()))
	}
	templateExcelPath := filepath.Join(consts.ProjectPath, consts.EmptyTemplatesXlsx)
	// templateExcelPath := filepath.Join(consts.DeployPath, "config", "local", "template.xlsx")
	tplF, err := os.Open(templateExcelPath)
	if err != nil {
		return "", errs.New(errcode.ErrServerLogic, fmt.Sprintf("读取模板文件错误: %s", err.Error()))
	}
	defer tplF.Close()
	newFile, err := os.CreateTemp("", fileName+"-*.xlsx")
	if err != nil {
		return "", errs.New(errcode.ErrServerLogic, fmt.Sprintf("创建临时文件错误: %s", err.Error()))
	}
	_, err = io.Copy(newFile, tplF)
	if err != nil {
		return "", errs.New(errcode.ErrServerLogic, fmt.Sprintf("拷贝模板文件错误: %s", err.Error()))
	}
	err = newFile.Close()
	if err != nil {
		return "", errs.New(errcode.ErrServerLogic, fmt.Sprintf("关闭临时文件错误: %s", err.Error()))
	}
	xlsxFile, err := excelize.OpenFile(newFile.Name())
	if err != nil {
		return "", errs.New(errcode.ErrServerLogic, fmt.Sprintf("读取临时文件错误: %s", err.Error()))
	}
	defer xlsxFile.Close()
	// 2. 遍历测点，分采集点和表达式点
	err = writeDriveInfo(xlsxFile, jsonContent.DrvInfo)
	if err != nil {
		log.Errorf("写入驱动信息错误: %s", err.Error())
		return "", err
	}
	err = writeExcelPoints(xlsxFile, jsonContent.PointsInfo)
	if err != nil {
		log.Errorf("写入测点信息错误: %s", err.Error())
		return "", err
	}
	err = xlsxFile.Save()
	if err != nil {
		log.Errorf("保存临时excel文件错误: %s", err.Error())
		return "", errs.New(errcode.ErrServerLogic, fmt.Sprintf("保存临时文件错误: %s", err.Error()))
	}
	return newFile.Name(), nil
}

// writeDriveInfo 写入"设备信息"sheet
func writeDriveInfo(file *excelize.File, driveInfo cmodel.DriverInfo) (err error) {
	if file == nil {
		return fmt.Errorf("excel文件为空指针")
	}
	err = file.SetCellValue(V2SheetDeviceInfoName, "A2", driveInfo.Class)
	if err != nil {
		return fmt.Errorf("设置单元格失败: %s", err.Error())
	}
	err = file.SetCellValue(V2SheetDeviceInfoName, "B2", driveInfo.Vendor)
	if err != nil {
		return fmt.Errorf("设置单元格失败: %s", err.Error())
	}
	err = file.SetCellValue(V2SheetDeviceInfoName, "C2", "")
	if err != nil {
		return fmt.Errorf("设置单元格失败: %s", err.Error())
	}
	err = file.SetCellValue(V2SheetDeviceInfoName, "D2", driveInfo.DriverName)
	if err != nil {
		return fmt.Errorf("设置单元格失败: %s", err.Error())
	}
	err = file.SetCellValue(V2SheetDeviceInfoName, "E2", driveInfo.ProtocolVersion)
	if err != nil {
		return fmt.Errorf("设置单元格失败: %s", err.Error())
	}
	err = file.SetCellValue(V2SheetDeviceInfoName, "F2", driveInfo.Extend)
	if err != nil {
		return fmt.Errorf("设置单元格失败: %s", err.Error())
	}
	return nil
}

// writeExcelPoints 写入"通讯点表"和"表达式计算"sheet
func writeExcelPoints(xlsxFile *excelize.File, points cmodel.InstancePointsInfo) (err error) {
	if xlsxFile == nil {
		return fmt.Errorf("excel文件为空指针")
	}
	collectPoints := []CollectTemplatePointModel{}
	expressPoints := []ExpressionPointModel{}
	for _, p := range points {
		unit, scale, desc := getCommonPointValueInfo(p)
		if p.ExprDef.Expr == "" {
			// 采集测点
			var cmd string
			if len(p.ProtocolDef.Command) > 2 {
				cmd = p.ProtocolDef.Command[:2]
			}
			collectPoints = append(collectPoints, CollectTemplatePointModel{
				SubDevice:             p.SubDevice,
				SignIdentifier:        string(p.ID),
				SignName:              p.Name,
				ValueType:             ConvertDataType(p.ValueType),
				ValueRW:               RwTransMap[p.Access],
				ValueUnit:             unit,
				ValueDesc:             desc,
				ValueRange:            p.ValueRange,
				ValueDeadZone:         p.ValueDeadZone,
				Cmd:                   cmd,
				Address:               p.ProtocolDef.Register,
				DataType:              p.ProtocolDef.Datatype,
				ByteOrder:             p.ProtocolDef.Byteorder,
				Scale:                 scale, // p.ProtocolDef.Scale字段是空的，是存在p.ValueDef中的
				Offset:                p.ProtocolDef.Offset,
				ExtendFunc:            p.ProtocolDef.Extend,
				CmdSendIntervalWeight: p.ProtocolDef.CmdIntervalWeight,
				IsNorthDefinition:     p.IsNorthDef,
				SimulationRule:        formatSimulationRuleStr(p.SimulatorDef),
			})
		} else {
			// 表达式测点
			expressPoints = append(expressPoints, ExpressionPointModel{
				SubDevice:         p.SubDevice,
				SignIdentifier:    string(p.ID),
				SignName:          p.Name,
				ValueType:         ConvertDataType(p.ValueType),
				ValueRW:           RwTransMap[p.Access],
				ValueUnit:         unit,
				ValueDesc:         desc,
				ValueRange:        p.ValueRange,
				ValueDeadZone:     p.ValueDeadZone,
				Expression:        p.ExprDef.Expr,
				ValueMap:          p.ExprDef.Mapping,
				IsNorthDefinition: p.IsNorthDef,
			})
		}
	}
	err = writeCollectPointsXlsx(xlsxFile, collectPoints)
	if err != nil {
		return
	}
	err = writeExpressPointsXlsx(xlsxFile, expressPoints)
	return
}

// writeCollectPointsXlsx 写"通讯点表"sheet
func writeCollectPointsXlsx(file *excelize.File, points []CollectTemplatePointModel) (err error) {
	if file == nil {
		return fmt.Errorf("excel文件为空指针")
	}
	startRow := pointValueIndexRow + 1
	for idx := startRow; idx < len(points)+startRow; idx++ {
		p := points[idx-startRow]
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("A%d", idx), p.SubDevice)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("B%d", idx), p.SignIdentifier)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("C%d", idx), p.SignName)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("D%d", idx), p.ValueType)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("E%d", idx), p.ValueRW)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("F%d", idx), p.ValueUnit)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("G%d", idx), p.ValueDesc)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("H%d", idx), p.ValueRange)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("I%d", idx), p.ValueDeadZone)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("J%d", idx), p.Cmd)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("K%d", idx), p.Address)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("L%d", idx), p.DataType)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("M%d", idx), p.ByteOrder)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("N%d", idx), p.Scale)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("O%d", idx), p.Offset)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("P%d", idx), p.ExtendFunc)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("Q%d", idx), p.CmdSendIntervalWeight)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("R%d", idx), p.IsNorthDefinition)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("S%d", idx), p.SimulationRule)
		_ = file.SetCellValue(V2SheetPointsName, fmt.Sprintf("T%d", idx), "") // 备注，无此字段
	}
	return nil
}

// writeExpressPointsXlsx 写"表达式计算"sheet
func writeExpressPointsXlsx(file *excelize.File, points []ExpressionPointModel) (err error) {
	if file == nil {
		return fmt.Errorf("excel文件为空指针")
	}
	startRow := expressionIndexRow + 1
	for idx := startRow; idx < len(points)+startRow; idx++ {
		p := points[idx-startRow]
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("A%d", idx), p.SubDevice)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("B%d", idx), p.SignIdentifier)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("C%d", idx), p.SignName)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("D%d", idx), p.ValueType)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("E%d", idx), p.ValueRW)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("F%d", idx), p.ValueUnit)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("G%d", idx), p.ValueDesc)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("H%d", idx), p.ValueRange)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("I%d", idx), p.ValueDeadZone)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("J%d", idx), p.Expression)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("K%d", idx), p.ValueMap)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("L%d", idx), p.IsNorthDefinition)
		_ = file.SetCellValue(V2SheetExpressionName, fmt.Sprintf("M%d", idx), "") // 备注，无此字段
	}
	return nil
}

// formatSimulationRuleStr 构造“通讯点表”的“模拟规则”列
func formatSimulationRuleStr(rule interface{}) string {
	input := make(map[string]string)
	tmp, ok := rule.(map[string]interface{})
	if !ok {
		log.Errorf("input simulation rule is invalid, format is %v", tmp)
	}
	for k, v := range tmp {
		if res, ok := v.(string); ok {
			input[k] = res
		}
	}
	if len(input) == 0 {
		return ""
	}
	funcName, ok := input["name"]
	if !ok {
		log.Errorf("input simulation rule is invalid, name is empty")
		return ""
	}
	switch funcName {
	case simulationFunctionStatic:
		return fmt.Sprintf("static(%s)", input["value"])
	case simulationFunctionRandom:
		return fmt.Sprintf("random(%s,%s,%s)", input["value"], input["min"], input["max"])
	case simulationFunctionMonotone:
		return fmt.Sprintf("monotone(%s,%s,%s)", input["min"], input["max"], input["step"])
	}
	return ""
}

// getCommonPointValueInfo 获取测点值的描述信息
func getCommonPointValueInfo(point cmodel.TemplateInstancePointInfo) (unit, scale, valDesc string) {
	if point.ValueDef == nil {
		return
	}
	// 第一种格式, {"unit":A, "scale":0.1, "valdesc": ""}
	def, ok := point.ValueDef.(map[string]interface{})
	if !ok {
		return
	}
	if _, ok = def["unit"]; ok {
		unit = def["unit"].(string)
	}
	if _, ok = def["scale"]; ok {
		scale = def["scale"].(string)
	}
	if _, ok = def["valdesc"]; ok {
		valDesc = def["valdesc"].(string)
	}

	// 第二种格式, {"val0":电流, "val1":电压} => "0=电流,1=电压"
	if _, ok = def["val0"]; ok {
		parts := make([]string, 0)
		idx := 0
		exist := true
		for exist {
			key := fmt.Sprintf("val%d", idx)
			if _, ok = def[key]; ok {
				parts = append(parts, fmt.Sprintf("%d=%s", idx, def[key]))
				idx++
			} else {
				exist = false
			}
		}
		valDesc = strings.Join(parts, ",")
	}
	return
}
