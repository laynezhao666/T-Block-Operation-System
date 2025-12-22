// Package influxdb provides influxdb v1 client
package influxdb

import (
	"context"
	"fmt"
	influxdbCli "github.com/influxdata/influxdb1-client/v2"
	"net"
	"sync"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/log"
	"trpc.group/trpc-go/trpc-go/naming/selector"
	"trpc.group/trpc-go/trpc-go/transport"
	dsn "trpc.group/trpc-go/trpc-selector-dsn"
)

func init() {
	// register client transport for trpc-selector
	transport.RegisterClientTransport(databaseName, &clientTransport{})
	selector.Register("influxdb+polaris", dsn.NewResolvableSelector("polaris", &URIHostExtractor{}))
	selector.Register("influxdb", dsn.DefaultSelector)
}

// clientTransport client transport for trpc-go client.Invoke
type clientTransport struct {
	client sync.Map // clientName -> influxdbClient
}

// influxdbClient influxdb client
type influxdbClient struct {
	influxdbCli.Client
	ip *net.TCPAddr
}

// RoundTrip Response to ctx inside, no need to return rspbuf here
func (ct *clientTransport) RoundTrip(ctx context.Context, _ []byte,
	callOpts ...transport.RoundTripOption) ([]byte, error) {
	msg := codec.Message(ctx)

	// get request body
	req, ok := msg.ClientReqHead().(*Request)
	if !ok {
		return nil, errs.NewFrameError(errs.RetClientEncodeFail,
			fmt.Sprintf("client transport: ReqHead should be type of *Request, current is %T", msg.ClientReqHead()))
	}

	opts := &transport.RoundTripOptions{}
	for _, o := range callOpts {
		o(opts)
	}

	// get parse address to get influxdb client
	cli, err := ct.getInfluxDBClient(msg.CalleeServiceName(), opts.Address)
	if err != nil {
		return nil, errs.WrapFrameError(err, errs.RetClientNetErr, "client transport get influxdb client error")
	}
	msg.WithRemoteAddr(cli.ip)

	// handle request
	switch req.Op {
	case OpPing:
		rsp, ok := msg.ClientRspHead().(*PingResponse)
		if !ok {
			return nil, errs.NewFrameError(errs.RetClientEncodeFail,
				fmt.Sprintf("client transport: RspHead should be type of *Response, current is %T",
					msg.ClientRspHead()))
		}
		err = ping(cli.Client, req, rsp)
	case OpQuery:
		rsp, ok := msg.ClientRspHead().(*Response)
		if !ok {
			return nil, errs.NewFrameError(errs.RetClientEncodeFail,
				fmt.Sprintf("client transport: RspHead should be type of *Response, current is %T",
					msg.ClientRspHead()))
		}
		err = query(cli.Client, req, rsp)
	case OpQueryAsChunk:
		rsp, ok := msg.ClientRspHead().(*ChunkedResponse)
		if !ok {
			return nil, errs.NewFrameError(errs.RetClientEncodeFail,
				fmt.Sprintf("client transport: RspHead should be type of *ChunkedResponse, current is %T",
					msg.ClientRspHead()))
		}
		err = queryAsChunk(cli.Client, req, rsp)
	case OpWrite:
		err = write(cli, req)
	case OpClose:
		log.Infof("close client:%s", cli.ip.String())
		err = cli.Client.Close()
	default:
		return nil, errs.NewFrameError(errs.RetServerSystemErr, "client transport invalid Request operation")
	}
	if err != nil {
		return nil, errs.WrapFrameError(err, errs.RetServerSystemErr, "client transport influxdb Response error")
	}

	return nil, nil
}

// getInfluxDBClient get influxdbCli.cli
func (ct *clientTransport) getInfluxDBClient(clientName, address string) (*influxdbClient, error) {
	if v, ok := ct.client.Load(clientName); ok {
		return v.(*influxdbClient), nil
	}

	if address == "" {
		return nil, fmt.Errorf("invalid %s address format, should be ip:port", databaseName)
	}

	// parse influxdb config
	userConfig, err := ParseAddress(address)
	if err != nil {
		return nil, err
	}

	// validate address
	addr, err := net.ResolveTCPAddr("tcp", userConfig.Address)
	if err != nil {
		return nil, fmt.Errorf("invalid influxdb address:%s format, should be ip:port", userConfig.Address)
	}

	// build influxdb client config
	httpConfig := influxdbCli.HTTPConfig{
		Addr:               "http://" + userConfig.Address,
		Username:           userConfig.Username,
		Password:           userConfig.Password,
		Timeout:            userConfig.Timeout,
		InsecureSkipVerify: userConfig.InsecureSkipVerify,
		WriteEncoding:      influxdbCli.ContentEncoding(userConfig.WriteEncoding),
	}

	// build influxdb client
	cli, err := influxdbCli.NewHTTPClient(httpConfig)
	if err != nil {
		return nil, fmt.Errorf("influxdb.NewHTTPClient error:%v", err)
	}

	c := &influxdbClient{
		Client: cli,
		ip:     addr,
	}

	ct.client.Store(clientName, c)
	return c, nil
}

// ping execute ping command
func ping(hc influxdbCli.Client, req *Request, rsp *PingResponse) error {
	costTime, version, err := hc.Ping(req.Timeout)
	if err != nil {
		return err
	}
	rsp.CostTime = costTime
	rsp.Version = version
	return nil
}

// query execute query command
func query(hc influxdbCli.Client, req *Request, rsp *Response) error {
	q := influxdbCli.Query{
		Command:         req.Command,
		Database:        req.Database,
		RetentionPolicy: req.RetentionPolicy,
		Precision:       req.Precision,
		Parameters:      req.Parameters,
	}

	result, err := hc.Query(q)
	if err != nil {
		return err
	}

	rsp.Response = result
	return nil
}

// query execute query command as chunk
func queryAsChunk(hc influxdbCli.Client, req *Request, rsp *ChunkedResponse) error {
	q := influxdbCli.Query{
		Command:         req.Command,
		Database:        req.Database,
		RetentionPolicy: req.RetentionPolicy,
		Precision:       req.Precision,
		Parameters:      req.Parameters,
		Chunked:         true,
		ChunkSize:       req.ChunkSize,
	}

	result, err := hc.QueryAsChunk(q)
	if err != nil {
		return err
	}

	rsp.Response = result
	return nil
}

// write execute write command
func write(hc influxdbCli.Client, req *Request) error {
	cfg := influxdbCli.BatchPointsConfig{
		Precision:        req.Precision,
		Database:         req.Database,
		RetentionPolicy:  req.RetentionPolicy,
		WriteConsistency: req.WriteConsistency,
	}
	batchPoints, err := influxdbCli.NewBatchPoints(cfg)
	if err != nil {
		return err
	}

	batchPoints.AddPoints(req.Points)

	return hc.Write(batchPoints)
}
