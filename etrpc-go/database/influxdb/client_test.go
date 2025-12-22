// Package influxdb provides influxdb v1 client
package influxdb

import (
	"context"
	"fmt"
	"testing"
	"time"

	influxdbCli "github.com/influxdata/influxdb1-client/v2"
	"trpc.group/trpc-go/trpc-go/client"

	"github.com/smartystreets/goconvey/convey"
)

type mockClient struct {
	handle func(context.Context, interface{}, interface{}, ...client.Option) error
}

func (mc *mockClient) Invoke(ctx context.Context, reqbody interface{}, rspbody interface{}, opt ...client.Option) error {
	return mc.handle(ctx, reqbody, rspbody, opt...)
}

var (
	name = "trpc.influxdb.xxx.xxx"

	fakeError = fmt.Errorf("fake error")
	failCli   = &mockClient{
		func(ctx context.Context, _ interface{}, _ interface{}, _ ...client.Option) error {
			return fakeError
		},
	}
	okCli = &mockClient{
		handle: func(ctx context.Context, req interface{}, rspbody interface{}, _ ...client.Option) error {
			switch req.(*Request).Op {
			case OpPing:
				rspbody.(*PingResponse).CostTime = time.Second
				rspbody.(*PingResponse).Version = "1.0.0"
			case OpQuery:
				rspbody.(*Response).Response = new(influxdbCli.Response)
			case OpQueryAsChunk:
				rspbody.(*ChunkedResponse).Response = new(influxdbCli.ChunkedResponse)
			}
			return nil
		},
	}
)

func Test_NewClientProxy(t *testing.T) {
	convey.Convey("Test_NewClientProxy", t, func() {
		rawClient := NewClientProxy(name).(*cli)
		convey.So(rawClient.client, convey.ShouldResemble, client.DefaultClient)
		convey.So(rawClient.serviceName, convey.ShouldEqual, name)
		convey.So(len(rawClient.opts), convey.ShouldEqual, 2)
	})
}

func Test_cli_Ping(t *testing.T) {
	timeout := time.Second
	ctx := context.Background()
	convey.Convey("Test_cli_Ping", t, func() {
		convey.Convey("ping error", func() {
			c := NewClientProxy(name).(*cli)
			c.client = failCli
			_, err := c.Ping(ctx, timeout)
			convey.So(err.Error(), convey.ShouldEqual, "fake error")
		})
		convey.Convey("ping ok", func() {
			client.DefaultClient = okCli
			_, err := NewClientProxy(name).Ping(ctx, timeout)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_cli_Query(t *testing.T) {
	q := Query{}
	ctx := context.Background()
	convey.Convey("Test_cli_Query", t, func() {
		convey.Convey("query error", func() {
			c := NewClientProxy(name).(*cli)
			c.client = failCli
			_, err := c.Query(ctx, q)
			convey.So(err.Error(), convey.ShouldEqual, "fake error")
		})
		convey.Convey("query ok", func() {
			client.DefaultClient = okCli
			_, err := NewClientProxy(name).Query(ctx, q)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_cli_QueryAsChunk(t *testing.T) {
	cq := ChunkedQuery{}
	ctx := context.Background()
	convey.Convey("Test_cli_QueryAsChunk", t, func() {
		convey.Convey("query error", func() {
			c := NewClientProxy(name).(*cli)
			c.client = failCli
			_, err := c.QueryAsChunk(ctx, cq)
			convey.So(err.Error(), convey.ShouldEqual, "fake error")
		})
		convey.Convey("query ok", func() {
			c := NewClientProxy(name).(*cli)
			c.client = okCli
			_, err := c.QueryAsChunk(ctx, cq)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_cli_Write(t *testing.T) {
	bp := BatchPoints{}
	convey.Convey("Test_cli_Write", t, func() {
		convey.Convey("write error", func() {
			c := NewClientProxy(name).(*cli)
			c.client = failCli
			err := c.Write(context.Background(), bp)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(err.Error(), convey.ShouldEqual, "fake error")
		})
		convey.Convey("write ok", func() {
			c := NewClientProxy(name).(*cli)
			c.client = okCli
			err := c.Write(context.Background(), bp)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func Test_cli_Close(t *testing.T) {
	ctx := context.Background()
	convey.Convey("Test_cli_Close", t, func() {
		convey.Convey("write error", func() {
			c := NewClientProxy(name).(*cli)
			c.client = failCli
			err := c.Close(ctx)
			convey.So(err.Error(), convey.ShouldEqual, "fake error")
		})
		convey.Convey("close ok", func() {
			c := NewClientProxy(name).(*cli)
			c.client = okCli
			err := c.Close(ctx)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
