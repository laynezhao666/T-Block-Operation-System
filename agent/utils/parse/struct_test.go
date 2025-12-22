package parse

import (
	"fmt"
	"testing"

	"agent/utils/parse/test"
)

func TestStruct(t *testing.T) {
	var r1 test.Response1
	var r2 test.Response2
	r2.Code = 1
	r2.Message = "jjjj"
	err := Struct(&r1, r2)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("%v\t%v\n", r1.Code, r1.Message)

	var r3 test.Response3
	err = Struct(&r3, &r1)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v\t%v\n", r3.Code, r3.Message)
}
