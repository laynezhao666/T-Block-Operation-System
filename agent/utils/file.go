package utils

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopyDir 复制文件夹
func CopyDir(src string, dst string) error {
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, path[len(src):])
		if info.IsDir() {
			if err = os.MkdirAll(dstPath, os.ModePerm); err != nil {
				return err
			}
		} else {
			if err = CopyFile(path, dstPath); err != nil {
				return err
			}
		}
		return nil
	})
}

// CopyFile 复制文件
func CopyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil

}
