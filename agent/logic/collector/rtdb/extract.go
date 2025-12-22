// Package rtdb 测点实时数据库
package rtdb

import (
	"agent/logic/collector/rtdb/model"
)

// SegmentRealVirtual 将 points 拆分为真实测点与虚拟测点
func SegmentRealVirtual(points model.DataPoints) (model.DataPoints, model.DataPoints) {
	l := len(points)
	if l == 0 {
		return make(model.DataPoints, 0), make(model.DataPoints, 0)
	}

	i := 0
	j := l - 1

	// 每次循环时，要么 i 自增，要么 j 自减，两者不会同时发生
	// 因此结束循环时，必有 i == j + 1
	for i <= j {
		// 开始本次循环前，points[:i-1] 均为真实测点，points[j+1:] 均为虚拟测点
		if !points[i].Rtd.Virtual {
			i++
			continue
		}
		if points[j].Rtd.Virtual {
			j--
			continue
		}
		// 此处 points[i] 必为虚拟测点， points[j] 必为真实测点
		points[i], points[j] = points[j], points[i]
		// 此处 points[i] 必为真实测点， points[j] 必为虚拟测点

		// 结束本次循环时，points[:i] 均为真实测点，points[j:] 均为虚拟测点
	}
	return points[:i], points[i:]
}
