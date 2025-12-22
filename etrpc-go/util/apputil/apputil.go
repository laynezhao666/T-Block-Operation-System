// Package apputil provides ...
// @author: xincili
// -------------------------------------------
package apputil

import (
	"sync"
	"trpc.group/trpc-go/trpc-go"
)

var (
	once sync.Once
	app  *Application // 应用基础信息
)

// Application 应用基础信息
type Application struct {
	Namespace string
	Env       string
	Container string
	IP        string
}

// NewApplication 返回应用基础信息
//
//	@return *Application: 应用基础信息
func NewApplication() *Application {
	once.Do(func() {
		app = &Application{
			Namespace: trpc.GlobalConfig().Global.Namespace,
			Env:       trpc.GlobalConfig().Global.EnvName,
			Container: trpc.GlobalConfig().Global.ContainerName,
			IP:        trpc.GlobalConfig().Global.LocalIP,
		}
	})
	return app
}
