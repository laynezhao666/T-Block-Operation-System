// Package healthcheck  provide healthcheck relate function
package healthcheck

import (
	"context"
	"etrpc-go/client/gorm"
	"etrpc-go/client/influxdb"
	"etrpc-go/client/redis"
	"github.com/pkg/errors"
	"strings"
	"trpc.group/trpc-go/trpc-database/mongodb"
	"trpc.group/trpc-go/trpc-go"
)

// CheckClients 检查DB类Client是否能够正常连接
func CheckClients(config *trpc.Config) error {
	for _, client := range config.Client.Service {
		serviceName := client.ServiceName
		split := strings.Split(serviceName, ".")
		if len(split) < 2 {
			continue
		}
		clientType := strings.ToLower(split[1])
		switch clientType {
		case "mysql", "postgres", "clickhouse", "sqlite":
			// 对数据库类的client进行连接测试
			_, err := gorm.NewClientProxy(serviceName)
			if err != nil {
				return errors.Wrapf(err, "connect %s client for %s fail", clientType, serviceName)
			}
		case "redis":
			// 对redis相关client进行连接测试
			_, err := redis.NewClientProxy(serviceName)
			if err != nil {
				return errors.Wrapf(err, "connect %s client for %s fail", clientType, serviceName)
			}
		case "mongodb":
			// 对mongodb相关client进行连接测试
			mongo := mongodb.NewClientProxy(serviceName)
			// 创建session进行连通性测试
			if session, err := mongo.StartSession(context.Background()); err != nil {
				return errors.Wrapf(err, "create %s client for %s fail", clientType, serviceName)
			} else {
				session.EndSession(context.Background())
			}
		case "influxdb":
			// 对influxdb相关client进行连接测试
			// influxdb 创建会自动进行ping操作
			_, err := influxdb.NewClientProxy(serviceName)
			if err != nil {
				return errors.Wrapf(err, "connect %s client for %s fail", clientType, serviceName)
			}
		default:
			continue
		}
	}
	return nil
}
