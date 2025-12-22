package sysdio

import "agent/logic/collector/device/model"

// SysdioValParser is a struct for parsing system IO values
type SysdioValParser struct {
	Pin      string
	UnaryFun string
}

// NewSysdioValueParser creates a new instance of SysdioValParser
func NewSysdioValueParser(params *model.ValParseParams) *SysdioValParser {
	return &SysdioValParser{
		Pin:      params.DataAddr,
		UnaryFun: params.Extend,
	}
}
