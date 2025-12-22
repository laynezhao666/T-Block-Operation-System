// Package gorm simply gorm use
package gorm

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"strings"
	"sync"
	tgorm "trpc.group/trpc-go/trpc-database/gorm"
	"trpc.group/trpc-go/trpc-go/client"
)

var (
	gormClientMapLock sync.Mutex
	gormClientMap     = make(map[string]*gorm.DB)
)

// GetDB 获取数据库连接，忽略错误，方便链式调用，先从缓存取，没有再新建
func GetDB(name string, opts ...client.Option) *gorm.DB {
	if cli, ok := gormClientMap[name]; ok {
		return cli
	}
	cli, err := NewClientProxy(name, opts...)
	if err != nil {
		panic(err)
	}
	return cli
}

// NewClientProxy 创建数据库连接客户端，这里新增支持了sqlite
func NewClientProxy(name string, opts ...client.Option) (*gorm.DB, error) {
	gormClientMapLock.Lock()
	defer gormClientMapLock.Unlock()

	// 判断是否是sqlite, 以后可以支持其他DB
	splitServiceName := strings.Split(name, ".")
	if len(splitServiceName) >= 2 {
		dbEngineType := splitServiceName[1]
		switch dbEngineType {
		case "sqlite":
			connPool := tgorm.NewConnPool(name)
			cli, err := gorm.Open(&sqlite.Dialector{Conn: connPool}, &gorm.Config{})
			if err != nil {
				return nil, err
			}
			gormClientMap[name] = cli
			return cli, nil
		}
	}
	// 其他非sqlite数据库，走trpc-gorm提供的方法
	cli, err := tgorm.NewClientProxy(name, opts...)
	if err != nil {
		return cli, err
	}
	gormClientMap[name] = cli
	return cli, err
}
