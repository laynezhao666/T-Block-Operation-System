package utils

import (
	model2 "agent/entity/model/data"
	"strings"

	"trpc.group/trpc-go/trpc-go/log"
)

var AllVirtualMap = map[string]struct{}{
	"almcount":              {},
	"almste":                {},
	"commste":               {},
	"point_throughput":      {},
	"range_resp_time":       {},
	"total_resp_time":       {},
	"avg_resp_time":         {},
	"max_resp_time":         {},
	"min_resp_time":         {},
	"success_req_in_period": {},
	"total_req_in_period":   {},
	"success_req":           {},
	"minute_success_req":    {},
	"total_req":             {},
	"minute_total_req":      {},
	"interruption":          {},
	"timeout_req":           {},
	"tms_delay_count_30":    {},
	"tms_delay_count_60":    {},
	"qua_err_count":         {},
	"qua_origin_err_count":  {},
	"success_msg_req":       {},
	"total_msg_req":         {},
	"Comm":                  {},
	"Comm_1":                {},
	"Comm_2":                {},
}

var VendorVirtualMap = map[string]struct{}{
	"commste": {},
	"Comm":    {},
	"Comm_1":  {},
	"Comm_2":  {},
}

// FilterInvalidVirtualPoints 过滤要上报的虚拟点
func FilterInvalidVirtualPoints(points []model2.KafkaPoint, forVendor bool) []model2.KafkaPoint {
	var filtered []model2.KafkaPoint
	for _, p := range points {
		// 找到最后一个 '.' 的位置
		idx := strings.LastIndex(p.ID, ".")
		if idx == -1 || idx == len(p.ID)-1 {
			// 格式不正确，跳过
			continue
		}
		attr := p.ID[idx+1:]

		if forVendor {
			// 厂商通道仅提供通讯状态类型基础虚拟点
			if _, ok := VendorVirtualMap[attr]; ok {
				filtered = append(filtered, p)
				continue
			}
		} else {
			// 内部通道提供全量虚拟点
			if _, ok := AllVirtualMap[attr]; ok {
				filtered = append(filtered, p)
				continue
			}
		}

		// 特殊：不属于虚拟点的测点，作为异常打印出来
		if _, ok := AllVirtualMap[attr]; !ok {
			log.Errorf("Invalid Virtual Point:%+v", p)
		}
	}
	return filtered
}
