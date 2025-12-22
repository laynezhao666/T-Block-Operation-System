// Package distributor 数据分发器
package distributor

import (
	"agent/entity/model/data"
)

// Distributor 数据分发者
type Distributor interface {
	// Distribute 执行数据分发的具体流程，data 指针不为空
	Distribute(data *data.DataUnit, args ...interface{})
}
// Distributors 数据分发者列表
type Distributors []Distributor
// BatchDistribute 批量分发数据
func (ds *Distributors) BatchDistribute(data *data.DataUnit, args ...interface{}) {
	for _, dt := range *ds {
		go func(dist Distributor) {
			dist.Distribute(data, args...)
		}(dt)
	}
}
