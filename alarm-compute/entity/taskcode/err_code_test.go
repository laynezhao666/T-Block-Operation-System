package taskcode

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	errTest := NewErr(&PointSvcErr, "err_test")
	fmt.Println(errTest.Error())
	fmt.Println(errTest.JudgeErrType(&PointSvcErr))
}
