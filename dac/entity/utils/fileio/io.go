// Package fileio 提供文件读写工具，支持JSON和YAML格式。
package fileio

// JSON 全局JSON文件读写器实例
// YAML 全局YAML文件读写器实例
var (
	JSON = &jsonIO{}
	YAML = &yamlIO{}
)

// RW 文件读写接口，支持读取和写入操作
type RW interface {
	Read(name string, pointer interface{}) error
	Write(name string, data interface{}) error
}
