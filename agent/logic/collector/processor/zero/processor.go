package zero

import (
	"agent/entity/consts"
	"agent/entity/definition"
	"agent/logic/collector/rtdb/model"
)

type pointsType map[definition.DataPointIDType]struct{}

// Processor Zero数据处理器
type Processor struct {
	enable     bool
	threshold  float64
	zeroPoints pointsType
}

// NewProcessor 创建 Zero 数据处理器
func NewProcessor(threshold int) *Processor {
	c := new(Processor)
	c.enable = threshold > 0 && threshold <= 100
	c.threshold = float64(threshold) / 100.0
	c.zeroPoints = make(pointsType)
	return c
}

func (p *Processor) addPoint(id definition.DataPointIDType) {
	p.zeroPoints[id] = struct{}{}
}

func (p *Processor) deletePoint(id definition.DataPointIDType) {
	delete(p.zeroPoints, id)
}

func (p *Processor) size() int {
	return len(p.zeroPoints)
}

// Do 处理测点
func (p *Processor) Do(pointCount int, pointID definition.DataPointIDType, point *model.RTValue) {
	if p == nil || !p.enable || pointCount == 0 {
		return
	}

	isZero, err := point.Pv.IsZero()
	if err != nil {
		return
	}

	if isZero {
		p.addPoint(pointID)
	} else {
		p.deletePoint(pointID)
	}

	targetCount := int(p.threshold * float64(pointCount))
	zeroCount := p.size()
	if targetCount == 0 || zeroCount < targetCount {
		return
	}

	id := string(pointID)
	filterLog.Warnf(id, "too many zero points, "+
		"point id: %v, total count: %v, threshold: %v, target count: %v, zeor count: %v",
		id, pointCount, p.threshold, targetCount, zeroCount)
	point.Qua = consts.QualityTooManyZero
}
