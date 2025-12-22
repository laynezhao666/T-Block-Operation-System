package os

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// CopyFile 复制文件
func CopyFile(src, dst string, overwrite, errorIfExist bool) error {
	dstInfo, err := os.Stat(dst)
	if err == nil {
		if dstInfo.IsDir() {
			return fmt.Errorf("can not copy file to directory: %s", dst)
		}
		if !overwrite {
			if errorIfExist {
				return fmt.Errorf("file already exists: %s", dst)
			}
			return nil
		}
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat src \"%v\": %w", src, err)
	}
	if srcInfo.IsDir() {
		return fmt.Errorf("%v is a directory", src)
	}

	dstDir := filepath.Dir(dst)
	if err = os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir \"%v\" error: %w", dstDir, err)
	}
	tempFile, err := os.CreateTemp(dstDir, "go_*.temp")
	if err != nil {
		return fmt.Errorf("create temp file error: %w", err)
	}
	tempFileName := tempFile.Name()
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFileName)
	}()

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %v error: %w", src, err)
	}

	// 复制文件
	if _, err = io.CopyN(tempFile, srcFile, srcInfo.Size()); err != nil {
		return fmt.Errorf("copy %v to %v error: %w", src, tempFile, err)
	}

	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("close %v error: %w", tempFileName, err)
	}

	// 复制文件权限
	if err = os.Chmod(tempFileName, srcInfo.Mode()); err != nil {
		return fmt.Errorf("chmod %v error: %w", tempFileName, err)
	}

	// 移动文件
	if err = os.Rename(tempFileName, dst); err != nil {
		return fmt.Errorf("rename %v to %v error: %w", tempFile, dst, err)
	}

	return nil
}

// Rename 移动文件
func Rename(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	if !errors.Is(err, syscall.EXDEV) {
		return err
	}

	// 若因跨文件分区导致移动失败
	// 则通过复制 + 删除的方法移动文件
	if err = CopyFile(src, dst, true, false); err != nil {
		return err
	}

	return os.Remove(src)
}
