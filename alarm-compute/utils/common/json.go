package common

import (
	"encoding/json"
	"strings"
)

// JSONMarshalNoErr JSONMarshalNoErr
func JSONMarshalNoErr(v interface{}) string {
	bjson, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	// 不使用 string(bjson)，减少 copy，提高性能
	builder := strings.Builder{}
	builder.Write(bjson)
	return builder.String()
}
