package iec103_siemens

// ASDU类型常量
const (
	ASDU_GenericService   = 0x0A // 通用分类服务监视方向
	ASDU_DisturbanceTable = 0x17 // 扰动表
	ASDU_DisturbanceReady = 0x1A // 扰动数据传输准备就绪
)

// asduTypeMap ASDU类型到报文类型的直接映射（不依赖COT）
var asduTypeMap = map[byte]string{
	ASDU_DisturbanceTable: Fault_data, // 扰动表
	ASDU_DisturbanceReady: Fault_data, // 扰动数据传输准备就绪
}

// genericServiceCotMap 通用分类服务(0x0A)的COT到报文类型映射
var genericServiceCotMap = map[byte]string{
	COT_Active:    Spontaneous_communication, // 自发
	COT_Cycle:     Cyclic_measurement,        // 循环
	COT_Call:      Call_data,                 // 呼叫
	COT_QueryEnd:  Query_End,                 // 查询结束
	COT_Read:      Call_energy,               // 读取
	COT_WriteFail: Write_Rsp,                 // 写失败
	COT_Write:     Write_Rsp,                 // 写
	COT_Write_Ack: Write_Rsp,                 // 写确认
}

// ParseActiveReport 解析主动上报数据
func ParseActiveReport(data []byte) *ActiveReport {
	if len(data) < IEC103_HeadLen+IEC103_MinItemLen {
		return nil
	}

	// 检查报文长度是否完整
	if len(data) < IEC103_HeadLen {
		return nil
	}

	report := &ActiveReport{
		Data: data,
	}

	asduType := data[IEC103_AsduOffset]
	cot := data[IEC103_CotOffset]

	// 优先查找ASDU类型直接映射
	if reportType, ok := asduTypeMap[asduType]; ok {
		report.Type = reportType
		return report
	}

	// 通用分类服务根据COT进行映射
	if asduType == ASDU_GenericService {
		if reportType, ok := genericServiceCotMap[cot]; ok {
			report.Type = reportType
		} else {
			report.Type = Unknown
		}
		return report
	}

	// 未知的ASDU类型
	report.Type = Unknown
	return report
}
