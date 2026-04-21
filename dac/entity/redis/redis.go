// Package redis 提供Redis客户端初始化和分布式锁功能。
package redis

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"dac/entity/config"
	"dac/entity/consts"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/client"
)

// rs 全局分布式锁管理器
// cli 全局Redis客户端
var (
	rs  *redsync.Redsync
	cli *Client
)

// Client Redis客户端封装
type Client struct {
	redis.UniversalClient
}

// Mutex 分布式互斥锁封装
type Mutex struct {
	*redsync.Mutex
}

// GetClient 获取全局Redis客户端实例
func GetClient() *Client {
	return cli
}

// parseTargetAndCreateClient 解析 target 地址并创建 redis 客户端
// target 格式: [scheme://]address[/db]?mode=xxx&user=xxx&password=xxx&master=xxx
func parseTargetAndCreateClient(target string, defaultDB int) (redis.UniversalClient, error) {
	index := strings.Index(target, "//")
	if index >= 0 {
		target = target[index+2:]
	}

	// 处理 userinfo@host 格式，将用户名密码和地址分离
	var userInfo string
	if atIdx := strings.Index(target, "@"); atIdx >= 0 {
		userInfo = target[:atIdx]
		target = target[atIdx+1:]
	}

	address := target
	var params []string
	if index = strings.LastIndex(target, "?"); index >= 0 {
		address = target[:index]
		params = strings.Split(target[index+1:], "&")
	}

	db := defaultDB
	if idx := strings.LastIndex(address, "/"); idx >= 0 {
		if n, err := strconv.Atoi(address[idx+1:]); err == nil {
			db = n
			address = address[:idx]
		}
	}

	paramsMap := make(map[string]string)
	for _, p := range params {
		kv := strings.Split(p, "=")
		if len(kv) != 2 {
			continue
		}
		paramsMap[kv[0]] = kv[1]
	}

	// 从 userInfo 中解析用户名和密码（格式: user:password 或 :password）
	var parsedUser, parsedPassword string
	if userInfo != "" {
		parts := strings.SplitN(userInfo, ":", 2)
		if len(parts) == 2 {
			parsedUser = parts[0]
			parsedPassword = parts[1]
		} else {
			parsedUser = parts[0]
		}
	}

	if config.C.Debug {
		config.Log.Infof("redis params: %+v, address: %v", paramsMap, address)
	}

	var c redis.UniversalClient

	// 优先使用 URL 中的用户名密码，如果 params 中也有则 params 优先
	user := parsedUser
	password := parsedPassword
	if v, ok := paramsMap["user"]; ok {
		user = v
	}
	if v, ok := paramsMap["password"]; ok {
		password = v
	}

	mode := paramsMap["mode"]
	switch mode {
	case "sentinel":
		c = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    paramsMap["master"],
			SentinelAddrs: strings.Split(address, ","),
			Username:      user,
			Password:      password,
			DB:            db,
			PoolSize:      50,
			MinIdleConns:  10,
		})
	case "cluster":
		c = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        strings.Split(address, ","),
			Username:     user,
			Password:     password,
			PoolSize:     50,
			MinIdleConns: 10,
		})
	default:
		c = redis.NewClient(&redis.Options{
			Addr:         address,
			Username:     user,
			Password:     password,
			DB:           db,
			PoolSize:     50,
			MinIdleConns: 10,
		})
	}

	return c, nil
}

// getTestClient 创建测试用Redis客户端
func getTestClient(target string) (redis.UniversalClient, error) {
	return parseTargetAndCreateClient(target, 1)
}

// getClient 从trpc配置中获取Redis客户端
func getClient(clientName string) (redis.UniversalClient, error) {
	cfg := trpc.GlobalConfig()
	var srv *client.BackendConfig = nil
	for _, s := range cfg.Client.Service {
		if s.ServiceName != clientName {
			continue
		}
		srv = s
		break
	}
	if srv == nil {
		return nil, fmt.Errorf("not find %v in trpc config", clientName)
	}

	return parseTargetAndCreateClient(srv.Target, 0)
}

// GetMutex 创建分布式互斥锁
func GetMutex(name string, value string, ttl time.Duration) Mutex {
	m := rs.NewMutex(name, redsync.WithGenValueFunc(func() (string, error) {
		return value, nil
	}), redsync.WithExpiry(ttl))
	var mutex Mutex
	mutex.Mutex = m
	return mutex
}

// Init 初始化Redis客户端和分布式锁管理器
func Init() error {
	c, err := getClient(consts.ClientRedis)
	//c, err := getTestClient(consts.ClientRedisTest)
	if err != nil {
		return err
	}

	cli = &Client{}
	cli.UniversalClient = c

	pool := goredis.NewPool(c)
	rs = redsync.New(pool)

	return nil
}
