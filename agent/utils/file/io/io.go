// Package io 文件读写接口
package io

var (
	JSON = &jsonIO{}
)

// RW 读写接口
type RW interface {
	// Read 读取文件
	Read(name string, pointer interface{}) error
	// Write 写入文件
	Write(name string, data interface{}) error
}
