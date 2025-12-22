// Package snmp snmp协议相关
package snmp

import (
	"fmt"
	"agent/utils/osal"
	"strings"

	"trpc.group/trpc-go/trpc-go/log"

	"agent/entity/consts"
	"agent/entity/definition/datatype"

	"github.com/gosnmp/gosnmp"
)

const (
	logInterval = 1000

	notSupportedPoint = "-9999"
)

var (
	errorCount = 0
)

type filterLogKey struct {
	IP    string
	Port  uint16
	Error string
}

func parseVariable(pdu *gosnmp.SnmpPDU, variant *osal.Variant) error {
	switch pdu.Type {
	case gosnmp.Integer:
		if v, ok := pdu.Value.(int); ok {
			variant.SetValue(v)
		} else {
			return fmt.Errorf("convert to int error, pdu: %+v", pdu)
		}
	case gosnmp.Counter32, gosnmp.Gauge32:
		if v, ok := pdu.Value.(uint); ok {
			variant.SetValue(v)
		} else {
			return fmt.Errorf("convert to uint error, pdu: %+v", pdu)
		}
	case gosnmp.TimeTicks:
		if v, ok := pdu.Value.(uint32); ok {
			variant.SetValue(v)
		} else {
			return fmt.Errorf("convert to uint32 error, pdu: %+v", pdu)
		}
	case gosnmp.Counter64:
		if v, ok := pdu.Value.(uint64); ok {
			variant.SetValue(v)
		} else {
			return fmt.Errorf("convert to uint64 error, pdu: %+v", pdu)
		}
	case gosnmp.OctetString:
		if v, ok := pdu.Value.([]byte); ok {
			variant.SetValue(string(v))
		} else if v, ok := pdu.Value.(string); ok {
			variant.SetValue(v)
		} else {
			return fmt.Errorf("convert to string error, pdu: %+v", pdu)
		}
	case gosnmp.NoSuchObject, gosnmp.NoSuchInstance, gosnmp.Null:
		variant.SetValue(notSupportedPoint)
	default:
		return fmt.Errorf("unprocessed pdu type: %+v, pdu: %+v", int(pdu.Type), pdu)
	}

	return nil
}

func retryGetValues(target *gosnmp.GoSNMP, oids []string, dataMap map[string]gosnmp.SnmpPDU) {
	if len(oids) <= 0 {
		return
	}

	result, err := target.Get(oids)
	if err != nil {
		log.Errorf("retryGetValues error: %+v, ip: %v, port: %v", err, target.Target, target.Port)
		return
	}

	for _, v := range result.Variables {
		if len(v.Name) > 1 && v.Name[0] == '.' {
			dataMap[v.Name[1:]] = v
		} else {
			log.Errorf("wrong oid: %+v", v.Name)
		}
	}

	if result.Error == gosnmp.NoError {
		return
	}
	if result.ErrorIndex == 0 || int(result.ErrorIndex) > len(oids) || int(result.ErrorIndex) > len(result.Variables) {
		return
	}

	index := result.ErrorIndex - 1
	GetEntryManager().Log(target.Target, target.Port, result.Variables[index].Name, result.Error)
	newOIDs := make([]string, 0, len(oids)-1)
	newOIDs = append(newOIDs, oids[0:index]...)
	newOIDs = append(newOIDs, oids[index+1:]...)
	retryGetValues(target, newOIDs, dataMap)
}

func getValues(target *gosnmp.GoSNMP, oids []string, variants []*osal.Variant, quas []consts.Quality) consts.Quality {
	oidNum := len(oids)
	if oidNum == 0 {
		return consts.QualityOk
	}
	if target == nil || oidNum != len(variants) {
		return consts.QualityConfigError
	}

	dataMap := make(map[string]gosnmp.SnmpPDU)
	result, err := target.Get(oids)
	if err != nil {
		k := filterLogKey{
			IP:    target.Target,
			Port:  target.Port,
			Error: err.Error(),
		}
		// 避免记录重复日志
		filterLog.Errorf(k, "snmp get return error: %+v, ip: %v, port: %v, first oid: \"%+v\", timeout: %v",
			err, target.Target, target.Port, oids[0], target.Timeout,
		)
		if strings.Index(k.Error, "timeout") >= 0 {
			return consts.QualityCmdRespTimeout
		}
		return consts.QualityCmdRespError
	}
	for _, v := range result.Variables {
		if len(v.Name) > 1 && v.Name[0] == '.' {
			dataMap[v.Name[1:]] = v
		} else {
			log.Errorf("wrong oid: %+v", v.Name)
			return consts.QualityConfigError
		}
	}

	if result.Error != gosnmp.NoError && needRetry {
		// 错误索引从 1 开始
		if result.ErrorIndex > 0 && int(result.ErrorIndex) <= len(oids) {
			index := result.ErrorIndex - 1
			GetEntryManager().Log(target.Target, target.Port, result.Variables[index].Name, result.Error)
			newOIDs := make([]string, 0, len(oids)-1)
			newOIDs = append(newOIDs, oids[0:index]...)
			newOIDs = append(newOIDs, oids[index+1:]...)
			retryGetValues(target, newOIDs, dataMap)
		} else {
			log.Warnf(
				"error status: %+v, error index: %v, oid list len: %v", result.Error, result.ErrorIndex, len(oids),
			)
		}
	}

	for i, oid := range oids {
		pdu, ok := dataMap[oid]
		if !ok {
			k := filterLogKey{
				IP:    target.Target,
				Port:  target.Port,
				Error: "",
			}
			filterLog.Errorf(k, "not find %v in snmp results, ip: %v, port: %v", oid, target.Target, target.Port)
			// 返回数据中未包含该测点
			quas[i] = consts.QualityCmdRespError
			continue
		}
		if err = parseVariable(&pdu, variants[i]); err != nil {
			if errorCount%logInterval == 0 {
				log.Errorf("ip: %+v, port: %+v, parseVariable error: %+v", target.Target, target.Port, err)
			}
			errorCount += 1
		}
	}
	return consts.QualityOk
}

func setValue(target *gosnmp.GoSNMP, oid string, val *osal.Variant, dataType datatype.DataType) consts.Quality {
	if target == nil || val == nil {
		return consts.QualityUncertain
	}

	pdu := gosnmp.SnmpPDU{
		Name: "." + oid,
	}
	switch dataType {
	case datatype.Int8Type, datatype.Int16Type, datatype.Int32Type, datatype.Int64Type, datatype.IntType:
		pdu.Type = gosnmp.Integer
		if v, err := val.AsInt64(); err != nil {
			log.Warnf("oid: %v, variant: %v AsInt64 error", oid, *val)
			return consts.QualityValueTypeError
		} else {
			pdu.Value = v
		}
	case datatype.Uint8Type, datatype.Uint16Type, datatype.Uint32Type, datatype.Uint64Type, datatype.UintType:
		pdu.Type = gosnmp.Gauge32
		if v, err := val.AsUint64(); err != nil {
			log.Warnf("oid: %v, variant: %v AsUint64 error", oid, *val)
			return consts.QualityValueTypeError
		} else {
			pdu.Value = v
		}
	case datatype.StringType:
		pdu.Type = gosnmp.OctetString
		if v, err := val.AsString(); err != nil {
			log.Warnf("oid: %v, variant: %v AsString error", oid, *val)
			return consts.QualityValueTypeError
		} else {
			pdu.Value = v
		}
	default:
		log.Warnf("oid: %v, unsupported data type: %v", oid, dataType)
		return consts.QualityConfigError
	}
	_, err := target.Set([]gosnmp.SnmpPDU{pdu})
	if err != nil {
		log.Errorf("snmp set return error: %v, ip: %v, port: %v", err, target.Target, target.Port)
		return consts.QualityCmdRespError
	}
	return consts.QualityOk
}
