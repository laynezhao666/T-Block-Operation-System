// Package copyutil provides copy utility functions.
package copyutil

import (
	"encoding/json"
	"github.com/jinzhu/copier"
	"github.com/mitchellh/copystructure"
	"google.golang.org/protobuf/types/known/structpb"
)

// Clone golang clone object by deep copy
func Clone(source any) (any, error) {
	return copystructure.Copy(source)
}

// Copy deep copies the fields with same type and Name from source to dest
// params must be pointer, otherwise panic
func Copy(source any, dest any) error {
	return copier.CopyWithOption(dest, source, copier.Option{DeepCopy: true})
}

// CopyNoNil deep copies the fields with same type and Name from source to dest, ignore nil value.
func CopyNoNil(source any, dest any) error {
	return copier.CopyWithOption(dest, source, copier.Option{DeepCopy: true, IgnoreEmpty: true})
}

// CopyWithOption deep copies the fields with same type and Name from source to dest with copier option.
func CopyWithOption(source any, dest any, opt copier.Option) error {
	return copier.CopyWithOption(dest, source, opt)
}

// Convert converts the source to dest by json marshalling and unmarshalling, slower than Copy function
func Convert(source any, dest any) error {
	jsonData, err := json.Marshal(source)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(jsonData, dest); err != nil {
		return err
	}
	return nil
}

// ConvertToStruct converts the source to structpb.Struct by json marshalling and unmarshalling.
func ConvertToStruct(source any) (*structpb.Struct, error) {
	structData := &structpb.Struct{}
	if err := Convert(source, structData); err != nil {
		return nil, err
	}
	return structData, nil
}
