package build

import (
	"bytes"
	"common/entity/consts"
	"common/entity/model"
	"errors"
	"fmt"
	"github.com/xuri/excelize/v2"
	"strconv"
	"strings"
	"trpcprotocol/cmdb"
)

type FieldRule struct {
	FieldName string // 英文字段名
	FieldType string // 字段类型(string/int/float/bool等)
	Required  bool   // 是否必填
}

type FieldMapping map[string]FieldRule // 表头中文 -> 字段规则

// parseExcelWithMapping 增强版Excel解析
func parseExcelWithMapping(file *cmdb.ReqImportModel_ExcelFile, mapping FieldMapping, fileName string) ([]map[string]interface{}, []string, error) {
	// 将[]byte转换为bytes.Reader以符合io.Reader接口
	reader := bytes.NewReader(file.Content)
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, nil, fmt.Errorf("读取Sheet1失败: %v", err)
	}

	if len(rows) == 0 {
		return nil, nil, errors.New("excel文件为空")
	}

	headers := rows[0]
	var (
		table []map[string]interface{}
		errs  []string
	)

	// 异常处理每一行数据
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		rowData := make(map[string]interface{})
		rowValid := true

		for j := 0; j < len(headers); j++ {
			header := strings.TrimSpace(headers[j])
			if rule, ok := mapping[header]; ok {
				value := ""
				if j < len(row) {
					value = strings.TrimSpace(row[j])
				}
				// 非空校验
				if rule.Required && value == "" {
					errs = append(errs, fmt.Sprintf("%s 第%d行: 字段[%s]不能为空", fileName, i+1, header))
					rowValid = false
					break
				}

				// 类型转换与校验
				converted, err := convertValue(value, rule.FieldType)
				if err != nil {
					errs = append(errs, fmt.Sprintf("%s 第%d行: 字段[%s]值[%s]类型错误(需要%s类型)",
						fileName, i+1, header, value, rule.FieldType))
					rowValid = false
					break
				}

				rowData[rule.FieldName] = converted
			}
		}

		if rowValid {
			table = append(table, rowData)
		}
	}

	return table, errs, nil
}

// convertValue 值类型转换
func convertValue(value string, fieldType string) (interface{}, error) {
	switch fieldType {
	case "string":
		return value, nil
	case "int":
		if value == "" {
			return 0, nil
		}
		v, err := strconv.Atoi(value)
		return v, err
	case "float":
		if value == "" {
			return 0.0, nil
		}
		v, err := strconv.ParseFloat(value, 64)
		return v, err
	case "bool":
		if value == "" {
			return false, nil
		}
		v, err := strconv.ParseBool(value)
		return v, err
	default:
		return nil, fmt.Errorf("unsupported type: %s", fieldType)
	}
}

// parseDeviceEntityExcel 标准设备解析
func parseDeviceEntityExcel(req *cmdb.ReqImportModel, fileName string) ([]*model.DeviceEntity, []string, error) {
	// 标准设备字段映射关系
	mapping := FieldMapping{
		"设备ID":   {FieldName: "DeviceGid", FieldType: "string", Required: true},
		"设备编号":   {FieldName: "DeviceNumber", FieldType: "string", Required: true},
		"设备编号路由": {FieldName: "DeviceNumberRoute", FieldType: "string", Required: false},
		"设备编号展示": {FieldName: "DeviceNumberShow", FieldType: "string", Required: false},
		"设备名称":   {FieldName: "DeviceName", FieldType: "string", Required: true},
		"IDC区域":  {FieldName: "IdcArea", FieldType: "string", Required: false},
		"IDC房间":  {FieldName: "FuncRoom", FieldType: "string", Required: false},
		"父级设备编号": {FieldName: "ParentDeviceNumber", FieldType: "string", Required: false},
		"设备类型英文": {FieldName: "DeviceTypeEn", FieldType: "string", Required: true},
		"设备类型中文": {FieldName: "DeviceTypeZh", FieldType: "string", Required: true},
		"应用类型英文": {FieldName: "ApplicationTypeEn", FieldType: "string", Required: true},
		"应用类型中文": {FieldName: "ApplicationTypeZh", FieldType: "string", Required: true},
		"归属应用类型": {FieldName: "BelongApplicationTypeEn", FieldType: "string", Required: false},
	}

	table, errs, err := parseExcelWithMapping(req.DeviceEntity, mapping, fileName)
	if err != nil {
		return nil, nil, err
	}

	// 构造返回数据
	var entities []*model.DeviceEntity
	for _, row := range table {
		entity := &model.DeviceEntity{
			DeviceGid:               row["DeviceGid"].(string),
			DeviceNumber:            row["DeviceNumber"].(string),
			DeviceNumberRoute:       row["DeviceNumberRoute"].(string),
			DeviceNumberShow:        row["DeviceNumberShow"].(string),
			MozuId:                  req.MozuId,
			DeviceName:              row["DeviceName"].(string),
			IdcArea:                 row["IdcArea"].(string),
			FuncRoom:                row["FuncRoom"].(string),
			ParentDeviceNumber:      row["ParentDeviceNumber"].(string),
			DeviceTypeEn:            row["DeviceTypeEn"].(string),
			DeviceTypeZh:            row["DeviceTypeZh"].(string),
			ApplicationTypeEn:       row["ApplicationTypeEn"].(string),
			ApplicationTypeZh:       row["ApplicationTypeZh"].(string),
			BelongApplicationTypeEn: row["BelongApplicationTypeEn"].(string),
		}
		entities = append(entities, entity)
	}

	return entities, errs, nil
}

// parseDevicePointExcel 标准测点解析
func parseDevicePointExcel(req *cmdb.ReqImportModel, fileName string) ([]*model.DevicePoint, []string, error) {
	// 标准测点字段映射关系
	mapping := FieldMapping{
		"设备ID":          {FieldName: "DeviceGid", FieldType: "string", Required: true},
		"设备编号":          {FieldName: "DeviceNumber", FieldType: "string", Required: true},
		"所属采集器":         {FieldName: "BelongCollector", FieldType: "string", Required: false},
		"测点英文名称":        {FieldName: "PointNameEn", FieldType: "string", Required: true},
		"测点中文名称":        {FieldName: "PointNameZh", FieldType: "string", Required: true},
		"测点类型":          {FieldName: "PointCategory", FieldType: "int", Required: true},
		"测点读写类型":        {FieldName: "PointRw", FieldType: "string", Required: false},
		"测点级别":          {FieldName: "PointLevel", FieldType: "string", Required: false},
		"测点表达式":         {FieldName: "Expression", FieldType: "string", Required: true},
		"表达式变量映射(设备ID)": {FieldName: "ExpressionMap", FieldType: "string", Required: false},
		"表达式变量映射(设备编号)": {FieldName: "ExpressionMapZh", FieldType: "string", Required: false},
		"值类型":           {FieldName: "ValueType", FieldType: "string", Required: false},
		"值有效范围":         {FieldName: "ValueValidRange", FieldType: "string", Required: false},
		"值单位":           {FieldName: "ValueUnit", FieldType: "string", Required: false},
		"值精度":           {FieldName: "ValuePrecision", FieldType: "string", Required: false},
		"值枚举映射":         {FieldName: "ValueEnum", FieldType: "string", Required: false},
	}

	table, errs, err := parseExcelWithMapping(req.DevicePoint, mapping, fileName)
	if err != nil {
		return nil, nil, err
	}
	// 构造返回数据
	var points []*model.DevicePoint
	for _, row := range table {
		point := &model.DevicePoint{
			DeviceGid:       row["DeviceGid"].(string),
			DeviceNumber:    row["DeviceNumber"].(string),
			BelongCollector: row["BelongCollector"].(string),
			PointNameEn:     row["PointNameEn"].(string),
			PointNameZh:     row["PointNameZh"].(string),
			PointKey: fmt.Sprintf("%s%s%s",
				row["DeviceGid"].(string), consts.PointConcatSep, row["PointNameEn"].(string)),
			PointCategory:   int32(row["PointCategory"].(int)),
			PointRw:         row["PointRw"].(string),
			PointLevel:      row["PointLevel"].(string),
			Expression:      row["Expression"].(string),
			ExpressionMap:   row["ExpressionMap"].(string),
			ExpressionMapZh: row["ExpressionMapZh"].(string),
			ValueType:       row["ValueType"].(string),
			ValueValidRange: row["ValueValidRange"].(string),
			ValueUnit:       row["ValueUnit"].(string),
			ValuePrecision:  row["ValuePrecision"].(string),
			ValueEnum:       row["ValueEnum"].(string),
			MozuId:          req.MozuId,
		}
		points = append(points, point)
	}

	return points, errs, nil
}

// parseCollectorDeviceExcel 采集设备解析
func parseCollectorDeviceExcel(req *cmdb.ReqImportModel, fileName string) ([]*model.CollectorDevice, []string, error) {
	// 采集设备字段映射关系
	mapping := FieldMapping{
		"设备ID":   {FieldName: "DeviceGid", FieldType: "string", Required: true},
		"设备编号":   {FieldName: "DeviceNumber", FieldType: "string", Required: true},
		"设备SN":   {FieldName: "DeviceSn", FieldType: "string", Required: false},
		"设备代号":   {FieldName: "DeviceCode", FieldType: "string", Required: true},
		"设备名称":   {FieldName: "DeviceName", FieldType: "string", Required: true},
		"设备类型英文": {FieldName: "DeviceTypeEn", FieldType: "string", Required: true},
		"设备类型中文": {FieldName: "DeviceTypeZh", FieldType: "string", Required: true},
		"采集器类型":  {FieldName: "CollectorType", FieldType: "int", Required: true},
		"采集通道类型": {FieldName: "ChannelType", FieldType: "string", Required: true},
		"采集通道地址": {FieldName: "ChannelId", FieldType: "string", Required: true},
		"采集通道信息": {FieldName: "ChannelLink", FieldType: "string", Required: true},
		"模版名称":   {FieldName: "TemplateName", FieldType: "string", Required: false},
		"模版信息":   {FieldName: "TemplateInfo", FieldType: "string", Required: false},
		"所属采集器":  {FieldName: "ParentDeviceNumber", FieldType: "string", Required: false},
		"扩展信息":   {FieldName: "Extend", FieldType: "string", Required: false},
	}

	table, errs, err := parseExcelWithMapping(req.CollectorDevice, mapping, fileName)
	if err != nil {
		return nil, nil, err
	}

	// 构造返回数据
	var devices []*model.CollectorDevice
	for _, row := range table {
		device := &model.CollectorDevice{
			DeviceGid:          row["DeviceGid"].(string),
			DeviceNumber:       row["DeviceNumber"].(string),
			DeviceSn:           row["DeviceSn"].(string),
			DeviceCode:         row["DeviceCode"].(string),
			DeviceName:         row["DeviceName"].(string),
			DeviceTypeEn:       row["DeviceTypeEn"].(string),
			DeviceTypeZh:       row["DeviceTypeZh"].(string),
			CollectorType:      int32(row["CollectorType"].(int)),
			ChannelType:        row["ChannelType"].(string),
			ChannelId:          row["ChannelId"].(string),
			ChannelLink:        row["ChannelLink"].(string),
			TemplateName:       row["TemplateName"].(string),
			TemplateInfo:       row["TemplateInfo"].(string),
			ParentDeviceNumber: row["ParentDeviceNumber"].(string),
			Extend:             row["Extend"].(string),
			MozuId:             req.MozuId,
		}
		devices = append(devices, device)
	}

	return devices, errs, nil
}

// parseCollectorTemplateExcel 采集模版解析
func parseCollectorTemplateExcel(req *cmdb.ReqImportModel, fileName string) ([]*model.CollectorTemplate, []string, error) {
	// 采集模版字段映射关系
	mapping := FieldMapping{
		"模版名称":   {FieldName: "TemplateName", FieldType: "string", Required: true},
		"设备类型英文": {FieldName: "DeviceTypeEn", FieldType: "string", Required: true},
		"设备类型中文": {FieldName: "DeviceTypeZh", FieldType: "string", Required: true},
		"设备制造商":  {FieldName: "Manufacturer", FieldType: "string", Required: false},
		"设备型号":   {FieldName: "DeviceModelEn", FieldType: "string", Required: false},
		"采集协议":   {FieldName: "ProtocolType", FieldType: "string", Required: true},
		"采集协议版本": {FieldName: "ProtocolVersion", FieldType: "string", Required: false},
		"协议扩展信息": {FieldName: "ProtocolExtend", FieldType: "string", Required: false},
	}

	table, errs, err := parseExcelWithMapping(req.CollectorTemplate, mapping, fileName)
	if err != nil {
		return nil, nil, err
	}
	// 构造返回数据
	var templates []*model.CollectorTemplate
	for _, row := range table {
		template := &model.CollectorTemplate{
			TemplateName:    row["TemplateName"].(string),
			DeviceTypeEn:    row["DeviceTypeEn"].(string),
			DeviceTypeZh:    row["DeviceTypeZh"].(string),
			Manufacturer:    row["Manufacturer"].(string),
			DeviceModelEn:   row["DeviceModelEn"].(string),
			ProtocolType:    row["ProtocolType"].(string),
			ProtocolVersion: row["ProtocolVersion"].(string),
			ProtocolExtend:  row["ProtocolExtend"].(string),
			MozuId:          req.MozuId,
		}
		templates = append(templates, template)
	}

	return templates, errs, nil
}

// parseCollectorTemplatePointExcel 采集模版测点解析
func parseCollectorTemplatePointExcel(req *cmdb.ReqImportModel, fileName string) ([]*model.CollectorTemplatePoint, []string, error) {
	//	采集模版测点字段映射关系
	mapping := FieldMapping{
		"模版名称":    {FieldName: "TemplateName", FieldType: "string", Required: true},
		"子设备名称":   {FieldName: "SubDevice", FieldType: "string", Required: false},
		"测点名称英文":  {FieldName: "PointNameEn", FieldType: "string", Required: true},
		"测点名称中文":  {FieldName: "PointNameZh", FieldType: "string", Required: true},
		"测点类型":    {FieldName: "PointType", FieldType: "string", Required: true},
		"测点读写分类":  {FieldName: "PointRw", FieldType: "string", Required: false},
		"变化定义规则":  {FieldName: "DeltaDef", FieldType: "string", Required: false},
		"校验规则":    {FieldName: "VerifyDef", FieldType: "string", Required: false},
		"表达式定义规则": {FieldName: "ExpDef", FieldType: "string", Required: false},
		"协议定义规则":  {FieldName: "ProtDef", FieldType: "string", Required: false},
		"值定义规则":   {FieldName: "ValDef", FieldType: "string", Required: false},
		"模拟定义规则":  {FieldName: "Simulator", FieldType: "string", Required: false},
	}

	table, errs, err := parseExcelWithMapping(req.TemplatePoint, mapping, fileName)
	if err != nil {
		return nil, nil, err
	}
	// 构造返回数据
	var points []*model.CollectorTemplatePoint
	for _, row := range table {
		point := &model.CollectorTemplatePoint{
			TemplateName: row["TemplateName"].(string),
			SubDevice:    row["SubDevice"].(string),
			PointNameEn:  row["PointNameEn"].(string),
			PointNameZh:  row["PointNameZh"].(string),
			PointType:    row["PointType"].(string),
			PointRw:      row["PointRw"].(string),
			DeltaDef:     row["DeltaDef"].(string),
			VerifyDef:    row["VerifyDef"].(string),
			ExpDef:       row["ExpDef"].(string),
			ProtDef:      row["ProtDef"].(string),
			ValDef:       row["ValDef"].(string),
			Simulator:    row["Simulator"].(string),
			MozuId:       req.MozuId,
		}
		points = append(points, point)
	}

	return points, errs, nil
}

// parseAlarmStrategyExcel 告警策略解析
func parseAlarmStrategyExcel(req *cmdb.ReqImportModel, fileName string) ([]*model.AlarmStrategy, []string, error) {
	// 告警策略字段映射关系
	mapping := FieldMapping{
		"设备ID":      {FieldName: "DeviceGid", FieldType: "string", Required: true},
		"策略ID":      {FieldName: "Rid", FieldType: "int", Required: true},
		"策略版本":      {FieldName: "RidVersion", FieldType: "string", Required: true},
		"策略类型":      {FieldName: "RidType", FieldType: "int", Required: true},
		"告警名称":      {FieldName: "AlarmName", FieldType: "string", Required: true},
		"告警表达式":     {FieldName: "AlarmExpression", FieldType: "string", Required: true},
		"告警表达式(中文)": {FieldName: "AlarmExpressionStr", FieldType: "string", Required: true},
		"恢复表达式":     {FieldName: "RestoreExpression", FieldType: "string", Required: true},
		"恢复表达式(中文)": {FieldName: "RestoreExpressionStr", FieldType: "string", Required: true},
		"表达式映射":     {FieldName: "ExpressionMap", FieldType: "string", Required: true},
		"告警级别":      {FieldName: "AlarmLevel", FieldType: "string", Required: true},
		"告警内容模版":    {FieldName: "ContentTemplate", FieldType: "string", Required: true},
		"告警负责人":     {FieldName: "Owner", FieldType: "string", Required: true},
		"计算复杂度":     {FieldName: "ComputeCost", FieldType: "int", Required: true},
	}

	table, errs, err := parseExcelWithMapping(req.AlarmStrategy, mapping, fileName)
	if err != nil {
		return nil, nil, err
	}

	//	构造返回数据
	var strategies []*model.AlarmStrategy
	for _, row := range table {
		strategy := &model.AlarmStrategy{
			DeviceGid:            row["DeviceGid"].(string),
			Rid:                  int64(row["Rid"].(int)),
			RidVersion:           row["RidVersion"].(string),
			RidType:              int32(row["RidType"].(int)),
			AlarmName:            row["AlarmName"].(string),
			AlarmExpression:      row["AlarmExpression"].(string),
			AlarmExpressionStr:   row["AlarmExpressionStr"].(string),
			RestoreExpression:    row["RestoreExpression"].(string),
			RestoreExpressionStr: row["RestoreExpressionStr"].(string),
			ExpressionMap:        row["ExpressionMap"].(string),
			AlarmLevel:           row["AlarmLevel"].(string),
			ContentTemplate:      row["ContentTemplate"].(string),
			Owner:                row["Owner"].(string),
			ComputeCost:          int32(row["ComputeCost"].(int)),
			MozuId:               req.MozuId,
		}
		strategies = append(strategies, strategy)
	}

	return strategies, errs, nil
}
