package encoding

import (
	"fmt"
	"testing"
)

func TestMD5String(t *testing.T) {
	fmt.Println(MD5String("123"))
}

func TestMD5File(t *testing.T) {
	fmt.Println(MD5File("md5.go"))
}
