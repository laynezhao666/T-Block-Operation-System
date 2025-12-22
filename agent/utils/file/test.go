package file

import (
	"errors"
	"io/fs"
	"os"
)

// TestExist 判断文件是否存在
func TestExist(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
