// Package influxdb provides ...
package influxdb

import (
	"context"
	"etrpc-go/database/influxdb"
	"sync"
	"time"
	"trpc.group/trpc-go/trpc-go/client"
)

var (
	lock        sync.RWMutex
	databaseMap = make(map[string]influxdb.Client)
)

// GetClient get influxdb client
func GetClient(name string, opts ...client.Option) influxdb.Client {
	if cli, ok := databaseMap[name]; ok {
		return cli
	}
	cli, err := NewClientProxy(name, opts...)
	if err != nil {
		panic(err)
	}
	return cli
}

// NewClientProxy create influxdb client
func NewClientProxy(name string, opts ...client.Option) (influxdb.Client, error) {
	lock.Lock()
	defer lock.Unlock()
	cli := influxdb.NewClientProxy(name, opts...)
	_, err := cli.Ping(context.Background(), time.Second*5)
	if err != nil {
		return nil, err
	}
	databaseMap[name] = cli
	return cli, nil
}
