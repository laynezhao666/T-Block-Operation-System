package http

import (
	"fmt"
	"testing"
	"time"
)

func TestFormat(t *testing.T) {
	fmt.Println(FormatTime(time.Now()))
}
