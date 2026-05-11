// Package utils 工具类代码
package utils

import (
	"encoding/json"
	"fmt"
	"time"

	"agent/entity/definition"
	"agent/entity/kafka"
	"agent/logic/cm"
	utils2 "agent/utils"
)

// DistributorArgs 推送时的参数
type DistributorArgs struct {
	Time     time.Time
	Interval int
	DataType int
	MozuID   string // 模组ID
}

// GetSendTimeAndInterval 获取发送时间、间隔、数据类型
func GetSendTimeAndInterval(args []interface{}) (time.Time, int, int) {
	sendTime := utils2.GetNowUTCTime()
	interval := definition.DefaultInterval
	dataType := definition.KafkaDataTypeStd
	if len(args) > 0 {
		if a, ok := args[0].(*DistributorArgs); ok && a != nil {
			sendTime = a.Time
			interval = a.Interval
			dataType = a.DataType
		}
	}
	return sendTime, interval, dataType
}

// GetDataType 获取数据类型
func GetDataType(args []interface{}) int {
	dataType := definition.KafkaDataTypeStd
	if len(args) > 0 {
		if a, ok := args[0].(*DistributorArgs); ok && a != nil {
			dataType = a.DataType
		}
	}
	return dataType
}

// GetMozuID 获取模组ID
func GetMozuID(args []interface{}) string {
	if len(args) > 0 {
		if a, ok := args[0].(*DistributorArgs); ok && a != nil {
			return a.MozuID
		}
	}
	return ""
}

// IsDefaultInterval 是否为默认间隔
func IsDefaultInterval(interval int) bool {
	return interval == definition.DefaultInterval
}

// GetBoxGidByDeviceGid 获取方仓的gid
func GetBoxGidByDeviceGid(deviceGid definition.DeviceGidType) definition.DeviceGidType {
	// TODO
	return deviceGid
}

// GetMessageKey 获取消息key
func GetMessageKey(deviceGid definition.DeviceGidType, timestamp int64, interval int, dataType int, n int) string {
	k := kafka.KafkaKey{
		MozuID:    cm.Worker().GetDeviceMozuID(deviceGid),
		DeviceGiD: string(deviceGid),
		WorkerID:  utils2.WorkerID(),
		Seq:       utils2.GetNextSequenceNumber(),
		Timestamp: timestamp,
		Interval:  interval,
		PubMs:     time.Now().UnixMilli(),
		Type:      dataType,
		CiID:      cm.Worker().GetCollectorGidByGid(deviceGid),
		N:         n,
	}
	if deviceGid == definition.StdDevice {
		k.BalancerKey = fmt.Sprintf("%v", k.Seq)
	} else {
		k.BalancerKey = string(deviceGid)
	}
	//// for debug mozu id为空
	//if k.MozuID == "" {
	//	log.Warnf("MozuID is empty,%+v", k)
	//}
	b, _ := json.Marshal(k)
	return string(b)
}
