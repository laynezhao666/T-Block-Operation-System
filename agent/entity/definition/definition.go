// Package definition provides definition
package definition

import (
	"fmt"
	"strconv"
	"strings"

	"agent/entity/consts"

	"github.com/segmentio/kafka-go"
)

type KafkaWriterType = kafka.Writer
type KafkaReaderType = kafka.Reader
type KafkaMessageType = kafka.Message

// DeviceGidType 设备 Gid 类型
type DeviceGidType string
type DeviceGidArrType []DeviceGidType

// PointType 测点类型 业务、采集
type PointType int

const (
	CollectPointType PointType = iota
	StdPointType
)

// PointIDType 测点 Gid 类型
type PointIDType string

// DataPointIDType 实例化的测点 Gid 类型
type DataPointIDType string
type DataPointIDsType []DataPointIDType

// IDPair 测点 Gid
type IDPair struct {
	DeviceGid       DeviceGidType
	PointID         PointIDType
	PointInstanceID DataPointIDType
}

// FloatType 浮点类型别名
type FloatType = float32

// GenerateDataPointID 根据 device gid （或id）与 pointID 组合出实例化的测点ID
func GenerateDataPointID(deviceGiD interface{}, pointID PointIDType) DataPointIDType {
	return DataPointIDType(fmt.Sprintf("%v%v%v", deviceGiD, consts.DefaultIDSep, pointID))
}

// SplitDataPointID 从实例化的测点ID中解析出设备Gid、测点ID
func SplitDataPointID(pointID DataPointIDType) (DeviceGidType, PointIDType, error) {
	pos := strings.LastIndex(string(pointID), consts.DefaultIDSep)
	if pos < 0 {
		return "", "", fmt.Errorf("invalid point id: \"%v\"", pointID)
	}
	return DeviceGidType(pointID[:pos]), PointIDType(pointID[pos+1:]), nil
}

// GetPointNo 从point_id （1307381722815410277.AcBranchE_1）中获取 No(AcBranchE_1)
func (id DataPointIDType) GetPointNo() string {
	if len(id) > 0 {
		list := strings.Split(string(id), ".")
		if len(list) > 0 {
			return list[len(list)-1]
		}
	}
	return ""
}

// GetPointGid 从point_id （1307381722815410277.AcBranchE_1）中获取 Gid(1307381722815410277)
func (id DataPointIDType) GetPointGid() string {
	if len(id) > 0 {
		list := strings.Split(string(id), ".")
		if len(list) > 0 {
			return list[0]
		}
	}
	return ""
}

// AddOne 增加一个
func (gid DeviceGidType) AddOne() (DeviceGidType, error) {
	intGid, _ := strconv.Atoi(string(gid))
	uint64Gid := uint64(intGid)

	return DeviceGidType(fmt.Sprintf("%v", uint64Gid+1)), nil
}

// IsInt 是否是整形
func (gid DeviceGidType) IsInt() bool {
	_, err := strconv.Atoi(string(gid))
	if err != nil {
		return false
	}
	return true
}
