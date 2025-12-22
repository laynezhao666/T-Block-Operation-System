package parse

import (
	"encoding/json"
	"fmt"
	"testing"

	jsoniterPkg "github.com/json-iterator/go"
)

var (
	testData = []byte(`{"id":"123","J":-999,"method":"pppp","params":{"d":"base_d","p":{"v":"base_v","m":{"a":"1","b":"2","c":"3"}},"s":{"v":"subs_vvvvv","l":[{"d":"sub_d","p":{"v":"sub_v","m":{"x":"0","y":"9","z":"8"}},"a":"sub_a"}]}}}`)
)

var (
	jsoniter = jsoniterPkg.ConfigCompatibleWithStandardLibrary
)

type CommonHeader struct {
	ID string `json:"id"`
	J  int
}

type CommonRequest struct {
	CommonHeader
	Method string `json:"method"`
}

type Request struct {
	CommonRequest
	Params Info `json:"params"`
	// Params interface{} `json:"params"`
}

type Request2 struct {
	CommonRequest
	Params Info2 `json:"params"`
}

type CommonResponse struct {
	CommonHeader
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Response struct {
	CommonResponse
	Data interface{} `json:"data"`
}

type BaseInfo struct {
	D string `json:"d"`
	P struct {
		V string            `json:"v"`
		M map[string]string `json:"m"`
	} `json:"p"`
}

type SubInfo struct {
	BaseInfo
	A string `json:"a"`
}

type Subs struct {
	V string    `json:"v"`
	L []SubInfo `json:"l"`
}

type Info struct {
	BaseInfo
	Subs  Subs   `json:"s"`
	Dummy string `json:"dummy"`
	X     int    `json:"x"`
}

type Info2 struct {
	BaseInfo
	Subs  Subs   `json:"s"`
	Dummy string `json:"dummy"`
}

func TestJSON(t *testing.T) {
	var err error
	b := []byte(`{"id":"123","J":-999,"method":"pppp","params":{"d":"base_d","p":{"v":"base_v","m":{"a":"1","b":"2","c":"3"}},"s":{"v":"subs_vvvvv","l":[{},{"d":"sub_d","p":{"v":"sub_v","m":{"x":"0","y":"9","z":"8"}},"a":"sub_a"}]}}}`)

	var data Request
	if err = json.Unmarshal(b, &data); err != nil {
		t.Errorf("json.Unmarshal error: %v", err)
	}

	var dst Info
	if err = JSON(&dst, &data.Params); err != nil {
		t.Errorf("JSON error: %v", err)
	}
	fmt.Printf("%+v\n", dst)

	// data.Params = nil
	if err = JSON(&dst, data.Params); err != nil {
		t.Errorf("JSON error: %v", err)
	}

	b = []byte(`{"params":{}}`)
	if err = json.Unmarshal(b, &data); err != nil {
		t.Errorf("json.Unmarshal error: %v", err)
	}
	if err = JSON(&dst, data.Params); err != nil {
		t.Errorf("JSON error: %v", err)
	}

	b = []byte(`{"params":null}`)
	if err = json.Unmarshal(b, &data); err != nil {
		t.Errorf("json.Unmarshal error: %v", err)
	}
	if err = JSON(&dst, data.Params); err != nil {
		t.Errorf("JSON error: %v", err)
	}
}

func TestJSON2(t *testing.T) {
	b := []byte(`{"1":"2"}`)
	var err error
	var s interface{}
	var d map[string]string
	if err = json.Unmarshal(b, &s); err != nil {
		t.Errorf("json.Unmarshal error: %v", err)
	}
	fmt.Printf("type: %T\n", s)
	if err = JSON(&d, s); err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", d)

	b = []byte(`{}`)
	if err = json.Unmarshal(b, &s); err != nil {
		t.Errorf("json.Unmarshal error: %v", err)
	}
	fmt.Printf("type: %T\n", s)
	if err = JSON(&d, s); err != nil {
		t.Error(err)
	}
	fmt.Printf("%+v\n", d)
}

func BenchmarkJSON(b *testing.B) {
	var data Request

	var err error
	if err = json.Unmarshal(testData, &data); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var info Info
		if err = JSON(&info, &data.Params); err != nil {
			panic(err)
		}
		// fmt.Printf("%+v\n", info)
	}
}

func BenchmarkJSONiter(b *testing.B) {
	var data Request
	err := jsoniter.Unmarshal(testData, &data)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var info Info
		tempBytes, err := jsoniter.Marshal(data.Params)
		if err != nil {
			panic(err)
		}
		if err = jsoniter.Unmarshal(tempBytes, &info); err != nil {
			panic(err)
		}
	}
}

func BenchmarkStandardJSON(b *testing.B) {
	var data Request
	_ = json.Unmarshal(testData, &data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var info Info
		tempBytes, err := json.Marshal(&data.Params)
		if err != nil {
			panic(err)
		}
		if err = json.Unmarshal(tempBytes, &info); err != nil {
			panic(err)
		}
	}
}
