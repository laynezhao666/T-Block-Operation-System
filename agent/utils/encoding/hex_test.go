package encoding

import (
	"fmt"
	"testing"
)

func TestHexString(t *testing.T) {
	input := []byte{0, 248, 2, 158}
	out := ParseBytesToHex(input)
	fmt.Println(out)
}
