package fdm

import (
	"errors"
	"agent/logic/collector/device/model"
)

// FDMValueParser FDM气体探测器值解析器
type FDMValueParser struct {
	// Extend 扩展参数
	Extend string
}

// NewFDMValParser 创建FDM值解析器
func NewFDMValParser(params *model.ValParseParams) *FDMValueParser {
	return &FDMValueParser{}
}

// GetPointValParser 获取测点值解析器
func GetPointValParser(p *model.PointInfo) (*FDMValueParser, error) {
	if p == nil {
		return nil, errors.New("point is nil")
	}

	valueParser, ok := p.Attr.ValParser.(*FDMValueParser)
	if !ok || valueParser == nil {
		return nil, errors.New("fdm value parser is not configured")
	}
	return valueParser, nil
}
