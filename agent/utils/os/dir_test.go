package os

import (
	"fmt"
	"testing"
)

func TestMoveDirectory(t *testing.T) {
	fmt.Println(MoveDirectory("a", "b"))
}
