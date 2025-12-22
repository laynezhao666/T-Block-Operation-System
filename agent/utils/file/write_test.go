package file

import (
	"fmt"
	"testing"
)

// TestSyncWrite 测试文件写入
func TestSyncWrite(t *testing.T) {
	fmt.Println(SyncWrite("1.txt", []byte("fjlwkejlw")))
}
