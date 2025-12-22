package os

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func walkDirectory(src, dst string, walkFile func(string, string) error) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("src directory \"%v\" error: %w", src, err)
	}
	if !srcInfo.IsDir() {
		return errors.New("src is not a directory")
	}

	if err = os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("mkdir dst directory \"%v\" error: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read src directory \"%v\" error: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err = walkDirectory(srcPath, dstPath, walkFile); err != nil {
				return fmt.Errorf("walkDirectory \"%v\" to \"%v\" error: %w", srcPath, dstPath, err)
			}
		} else {
			if err = walkFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("walkFile file \"%v\" to \"%v\" error: %w", srcPath, dstPath, err)
			}
		}
	}

	return nil
}

// MoveDirectory 移动目录
func MoveDirectory(src, dst string) error {
	if err := walkDirectory(src, dst, Rename); err != nil {
		return err
	}

	return os.RemoveAll(src)
}

// CopyDirectory 复制目录
func CopyDirectory(src, dst string, overwrite, errorIsExist bool) error {
	return walkDirectory(src, dst, func(s string, d string) error {
		return CopyFile(s, d, overwrite, errorIsExist)
	})
}
