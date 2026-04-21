// Package config 提供门禁服务的全局配置和日志实例。
package config

import (
	"dac/entity/consts"
)

// Init 初始化全局配置，填充默认值
func Init() {
	C.initGIDMapping()
}

// fillURL 填充URL路径，若为空则使用默认值，并添加前缀
func fillURL(path *string, prefix, defaultPath string) {
	v := *path
	if len(v) == 0 {
		v = defaultPath
	}

	*path = prefix + v
}

// initGIDMapping 初始化GID映射相关的URL配置
func (c *Config) initGIDMapping() {
	if len(c.GIDMapping.Prefix) == 0 {
		c.GIDMapping.Prefix = consts.DefaultGIDMappingPrefix
	}

	fillURL(&c.GIDMapping.URL.Location,
		c.GIDMapping.Prefix,
		consts.DefaultGIDMappingURLLocation)
	fillURL(&c.GIDMapping.URL.ConvertCode,
		c.GIDMapping.Prefix,
		consts.DefaultGIDMappingURLConvertCode)
	fillURL(&c.GIDMapping.URL.Rooms,
		c.GIDMapping.Prefix,
		consts.DefaultGIDMappingURLRooms)
}
