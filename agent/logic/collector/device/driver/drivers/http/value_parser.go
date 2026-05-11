package http

import (
	"strings"

	"agent/logic/collector/device/model"
)

type ValueParser struct {
	Addr     string
	Extend   string
	DataType string
}

type HTTPValueParser struct {
	Keys []string
	Base ValueParser
}

func NewHTTPValueParser(params *model.ValParseParams) *HTTPValueParser {
	var dataAddr string
	if params != nil {
		dataAddr = params.DataAddr
	}
	return &HTTPValueParser{
		Keys: splitEscaped(dataAddr, '.', '\\'),
		// 同步保留附加元数据（若有）
		Base: ValueParser{
			Addr:     params.DataAddr,
			Extend:   params.Extend,
			DataType: params.DataType,
		},
	}
}

// 点位路径分割，支持反斜杠转义： a.b\.c.d  => ["a","b.c","d"]
func splitEscaped(s string, sep, esc rune) []string {
	var out []string
	var b strings.Builder
	escaped := false
	for _, r := range s {
		if escaped {
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == esc {
			escaped = true
			continue
		}
		if r == sep {
			out = append(out, b.String())
			b.Reset()
			continue
		}
		b.WriteRune(r)
	}
	out = append(out, b.String())
	return out
}
