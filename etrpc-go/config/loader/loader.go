// Package loader do config load logic
package loader

import (
	"etrpc-go/config/cache"
	"etrpc-go/config/loader/file"
	"etrpc-go/config/util"
	"etrpc-go/util/iputil"
)

// LoadConfig 加载配置项,出现错误直接panic
func LoadConfig() {
	// 加载本地文件配置
	fileCfg, err := file.Load(file.GetLocalConfigPath())
	if err != nil {
		panic(err)
	}
	setDefaultConfig(fileCfg)
	// 将配置中的变量进行替换
	if fileCfg, err = util.ExpandEnv(fileCfg); err != nil {
		panic(err)
	}
	if err = cache.RefreshConfig(fileCfg); err != nil {
		panic(err)
	}
}

// setDefaultConfig 用于在替换变量前设置默认值
//
//	@param map[string]any	已有配置项
func setDefaultConfig(cfg map[string]any) {
	util.SetDefaultIfAbsent(cfg, "etrpc.service_port", 8080)
	util.SetDefaultIfAbsent(cfg, "etrpc.trpc_service_port", 8081)
	// 本地启动，这些环境变量没有设置，设置默认值
	util.SetOSEnvIfAbsent("POD_IP", iputil.GetLocalIP())
	util.SetOSEnvIfAbsent("POD_NAME", "local-machine")
	util.SetOSEnvIfAbsent("TRPC_NAMESPACE", "Development")
}
