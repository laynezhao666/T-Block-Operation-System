// Package virtualpoints 提供门禁控制器虚拟测点的采集和上报功能。
package virtualpoints

import (
	"trpc.group/trpc-go/trpc-go/metrics"
)

// getAttrList 将属性map转换为metrics维度列表
func getAttrList(attrs map[string]string) []*metrics.Dimension {
	attrsList := make([]*metrics.Dimension, 0, len(attrs))
	for n, v := range attrs {
		attrsList = append(attrsList, &metrics.Dimension{
			Name:  n,
			Value: v,
		})
	}
	return attrsList
}
