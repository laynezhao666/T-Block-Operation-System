package utils

import (
	"fmt"
	"testing"
)

func TestToHex(t *testing.T) {
	buff := []byte{01, 02, 03}
	fmt.Println(ToHex(buff, " "))
}
