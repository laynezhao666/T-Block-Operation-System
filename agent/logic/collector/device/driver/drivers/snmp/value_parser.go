package snmp

import (
	"agent/entity/definition/datatype"
)

// SnmpValueParser snmp数据解析器
type SnmpValueParser struct {
	OID      string
	Extend   string
	DataType datatype.DataType
}
