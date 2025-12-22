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
}
// GetSendTimeAndInterval 获取发送时间和间隔
func GetSendTimeAndInterval(args []interface{}) (time.Time, int) {
	sendTime := utils2.GetNowUTCTime()
	interval := definition.DefaultInterval
	if len(args) > 0 {
		if a, ok := args[0].(*DistributorArgs); ok && a != nil {
			sendTime = a.Time
			interval = a.Interval
		}
	}
	return sendTime, interval
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
func GetMessageKey(deviceGid definition.DeviceGidType, timestamp int64, interval int) string {
	k := kafka.KafkaKey{
		MozuID:    cm.Worker().GetDeviceMozuID(deviceGid), // 标准点数据发送未按设备维度聚合，全园区平铺，故此字段为空
		DeviceGiD: string(deviceGid),
		WorkerID:  utils2.WorkerID(),
		Seq:       utils2.GetNextSequenceNumber(),
		Timestamp: timestamp,
		Interval:  interval,
		PubMs:     time.Now().UnixMilli(),
	}
	if deviceGid == definition.StdDevice {
		k.BalancerKey = fmt.Sprintf("%v", k.Seq)
	} else {
		k.BalancerKey = string(deviceGid)
	}
	b, _ := json.Marshal(k)
	return string(b)
}
