// Package influxdb provides influxdb v1 client
package influxdb

import (
	"context"
	"fmt"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	influxdbCli "github.com/influxdata/influxdb1-client/v2"
	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/codec"
)

const databaseName = "influxdb"

// OpEnum is the database operation enumeration type.
type OpEnum int

const (
	OpPing OpEnum = iota + 1
	OpQuery
	OpQueryAsChunk
	OpWrite
	OpClose
)

// String converts the operation type constant to human-readable text.
func (op OpEnum) String() string {
	return [...]string{
		"",
		"Ping",
		"Query",
		"QueryAsChunk",
		"Write",
		"Close",
	}[op]
}

// Client defines the influxdb client interface.
type Client interface {
	Ping(ctx context.Context, timeout time.Duration) (*PingResponse, error)
	Query(ctx context.Context, query Query) (*influxdbCli.Response, error)
	QueryAsChunk(ctx context.Context, query ChunkedQuery) (*influxdbCli.ChunkedResponse, error)
	Write(ctx context.Context, batchPoints BatchPoints) error
	Close(ctx context.Context) error
}

// cli influxdb client proxy struct
type cli struct {
	serviceName string
	client      client.Client
	opts        []client.Option
}

// NewClientProxy create influxdb client proxy: trpc.influxdb.xxx
func NewClientProxy(name string, opts ...client.Option) Client {
	c := &cli{
		serviceName: name,
		client:      client.DefaultClient,
	}
	c.opts = make([]client.Option, 0, len(opts)+2)
	c.opts = append(c.opts, opts...)
	c.opts = append(c.opts,
		client.WithProtocol(databaseName),
		client.WithDisableServiceRouter())
	return c
}

// Query influxdb query struct
type Query struct {
	// Command query Command: such as create Database, show Database, select ... and influxQL
	Command  string
	Database string
	// RetentionPolicy is the retention policy of the Points.
	RetentionPolicy string
	// Precision support ns(nanosecond, default) u(microsecond) ms(millisecond) s(second) m(minute) h(hour)
	Precision  string
	Parameters map[string]any
}

// ChunkedQuery chunked query struct
type ChunkedQuery struct {
	Query
	ChunkSize int
}

// BatchPoints batch Points for writing
type BatchPoints struct {
	// Should use influxdb.NewPoint() to create a point. GC unfriendly :(
	Points []*influxdbCli.Point
	// Database influxdb Database
	Database string
	// Precision is timestamp accuracy
	Precision string
	// RetentionPolicy is the retention policy of the Points.
	RetentionPolicy string
	// Write consistency is the number of servers required to confirm write.
	WriteConsistency string
}

// Request influxdb Request body
type Request struct {
	Command          string
	Database         string
	RetentionPolicy  string
	WriteConsistency string
	Precision        string
	Parameters       map[string]any
	ChunkSize        int
	Points           []*influxdbCli.Point
	Op               OpEnum
	Timeout          time.Duration
}

// Response wrap influxdb Response
type Response struct {
	Response *influxdbCli.Response
}

// ChunkedResponse wrap influxdb ChunkedResponse
type ChunkedResponse struct {
	Response *influxdbCli.ChunkedResponse
}

// PingResponse influxdb ping response
type PingResponse struct {
	CostTime time.Duration
	Version  string
}

// Ping influxdb ping
//
// @param ctx: 上下文信息
// @param timeout: 超时时间
// @return time.Duration: 请求耗时
// @return string: influxdb版本
// @return error: 错误信息
func (c *cli) Ping(ctx context.Context, timeout time.Duration) (*PingResponse, error) {
	req := &Request{
		Op:      OpPing,
		Timeout: timeout,
	}
	rsp := new(PingResponse)

	ctx, msg := codec.WithCloneMessage(ctx)
	msg.WithClientRPCName(fmt.Sprintf("/%s/Ping", c.serviceName))
	msg.WithCalleeServiceName(c.serviceName)
	msg.WithSerializationType(codec.SerializationTypeUnsupported)
	msg.WithCompressType(codec.CompressTypeNoop)
	msg.WithClientReqHead(req)
	msg.WithClientRspHead(rsp)

	err := c.client.Invoke(ctx, req, rsp, c.opts...)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// Query influxdb query Command
//
// @param ctx: 上下文信息
// @param query: 查询命令
// @return *influxdbCli.Response: 返回数据
// @return error: 错误信息
func (c *cli) Query(ctx context.Context, query Query) (*influxdbCli.Response, error) {
	req := &Request{
		Command:         query.Command,
		Database:        query.Database,
		RetentionPolicy: query.RetentionPolicy,
		Precision:       query.Precision,
		Parameters:      query.Parameters,
		Op:              OpQuery,
	}
	rsp := new(Response)

	ctx, msg := codec.WithCloneMessage(ctx)
	msg.WithClientRPCName(fmt.Sprintf("/%s/Query", c.serviceName))
	msg.WithCalleeServiceName(c.serviceName)
	msg.WithSerializationType(codec.SerializationTypeUnsupported)
	msg.WithCompressType(codec.CompressTypeNoop)
	msg.WithClientReqHead(req)
	msg.WithClientRspHead(rsp)

	err := c.client.Invoke(ctx, req, rsp, c.opts...)
	if err != nil {
		return nil, err
	}

	return rsp.Response, nil
}

// QueryAsChunk influxdb chunked query
//
// @param ctx: 上下文信息
// @param query: 查询命令
// @return *influxdbCli.ChunkedResponse: 返回数据
// @return error: 错误信息
func (c *cli) QueryAsChunk(ctx context.Context, query ChunkedQuery) (*influxdbCli.ChunkedResponse, error) {
	req := &Request{
		Command:         query.Command,
		Database:        query.Database,
		RetentionPolicy: query.RetentionPolicy,
		Precision:       query.Precision,
		Parameters:      query.Parameters,
		ChunkSize:       query.ChunkSize,
		Op:              OpQueryAsChunk,
	}
	rsp := new(ChunkedResponse)

	ctx, msg := codec.WithCloneMessage(ctx)
	defer codec.PutBackMessage(msg)
	msg.WithClientRPCName(fmt.Sprintf("/%s/QueryAsChunk", c.serviceName))
	msg.WithCalleeServiceName(c.serviceName)
	msg.WithSerializationType(codec.SerializationTypeUnsupported)
	msg.WithCompressType(codec.CompressTypeNoop)
	msg.WithClientReqHead(req)
	msg.WithClientRspHead(rsp)

	err := c.client.Invoke(ctx, req, rsp, c.opts...)
	if err != nil {
		return nil, err
	}

	return rsp.Response, nil
}

// Write influxdb write Command
//
// @param ctx: 上下文信息
// @param batchPoints: 批量数据
// @return error: 错误信息
func (c *cli) Write(ctx context.Context, batchPoints BatchPoints) error {
	req := &Request{
		Database:         batchPoints.Database,
		RetentionPolicy:  batchPoints.RetentionPolicy,
		Precision:        batchPoints.Precision,
		WriteConsistency: batchPoints.WriteConsistency,
		Points:           batchPoints.Points,
		Op:               OpWrite,
	}

	ctx, msg := codec.WithCloneMessage(ctx)
	defer codec.PutBackMessage(msg)
	msg.WithClientRPCName(fmt.Sprintf("/%s/Write", c.serviceName))
	msg.WithCalleeServiceName(c.serviceName)
	msg.WithSerializationType(codec.SerializationTypeUnsupported)
	msg.WithCompressType(codec.CompressTypeNoop)
	msg.WithClientReqHead(req)

	return c.client.Invoke(ctx, req, nil, c.opts...)
}

// Close influxdb close Command
//
// @param ctx: 上下文信息
// @return error: 错误信息
func (c *cli) Close(ctx context.Context) error {
	req := &Request{
		Op: OpClose,
	}

	ctx, msg := codec.WithCloneMessage(ctx)
	defer codec.PutBackMessage(msg)
	msg.WithClientRPCName(fmt.Sprintf("/%s/Close", c.serviceName))
	msg.WithCalleeServiceName(c.serviceName)
	msg.WithSerializationType(codec.SerializationTypeUnsupported)
	msg.WithCompressType(codec.CompressTypeNoop)
	msg.WithClientReqHead(req)

	return c.client.Invoke(ctx, req, nil, c.opts...)
}
