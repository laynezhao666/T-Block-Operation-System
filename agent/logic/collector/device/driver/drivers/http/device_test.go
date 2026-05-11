package http

import (
	"reflect"
	"testing"
)

// Test_getByKeys 测试getByKeys
func Test_getByKeys(t *testing.T) {
	type args struct {
		root    any
		keys    []string
		quaName string
	}
	tests := []struct {
		name  string
		args  args
		want  any
		want1 bool
		want2 string
	}{
		{"getByKeys1", args{root: map[string]any{"data": map[string]any{"BPC_1.DcUin": map[string]any{"pv": "1", "qua": "-422"}}},
			keys: []string{"data", "BPC_1.DcUin", "pv"}, quaName: ""}, "1", true, "0"},
		{"getByKeys2", args{root: map[string]any{"data": map[string]any{"BPC_1.DcUin": map[string]any{"pv": "1", "qua": "-422"}}},
			keys: []string{"data", "BPC_1.DcUin", "pv"}, quaName: "qua"}, "1", true, "-422"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, qua, got1 := getByKeys(tt.args.root, tt.args.keys, tt.args.quaName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getByKeys() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getByKeys() got1 = %v, want %v", got1, tt.want1)
			}
			if qua != tt.want2 {
				t.Errorf("getByKeys() qua = %v, want %v", qua, tt.want2)
			}
		})
	}
}
