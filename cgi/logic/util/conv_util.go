package util

import (
	"encoding/json"

	structpb "github.com/golang/protobuf/ptypes/struct"
)

// JsonStrToPbStruct 将Json字符串直接转化为structpb.Struct
func JsonStrToPbStruct(jsonStr string) *structpb.Struct {
	res := &structpb.Struct{}
	if len(jsonStr) > 0 {
		_ = json.Unmarshal([]byte(jsonStr), res)
	}
	return res
}
