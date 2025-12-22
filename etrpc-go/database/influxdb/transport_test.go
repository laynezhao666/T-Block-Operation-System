// Package influxdb provides influxdb v1 client
package influxdb

import (
	"context"
	"github.com/agiledragon/gomonkey/v2"
	"reflect"
	"testing"
	"time"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/transport"

	influxdbCli "github.com/influxdata/influxdb1-client/v2"
	"github.com/smartystreets/goconvey/convey"
)

var (
	tc           = &clientTransport{}
	emptyAddress = ""
	validAddress = "user:passwd@127.0.0.1:8086"
)

type fakeClient struct{}

func (c *fakeClient) Ping(timeout time.Duration) (time.Duration, string, error) {
	return 0, "1.0.0", nil
}

func (c *fakeClient) Write(bp influxdbCli.BatchPoints) error {
	return nil
}

func (c *fakeClient) Query(q influxdbCli.Query) (*influxdbCli.Response, error) {
	return new(influxdbCli.Response), nil
}

func (c *fakeClient) QueryAsChunk(q influxdbCli.Query) (*influxdbCli.ChunkedResponse, error) {
	return new(influxdbCli.ChunkedResponse), nil
}

func (c *fakeClient) Close() error {
	return nil
}

func TestClientTransport_RoundTrip(t *testing.T) {
	convey.Convey("TestClientTransport_RoundTrip", t, func() {
		convey.Convey("ClientReqHead not *Request", func() {
			ctx := context.Background()
			_, err := tc.RoundTrip(ctx, nil)
			convey.So(err.Error(), convey.ShouldContainSubstring, "client transport: ReqHead should be type of *Request")
		})
		ctx, msg := codec.WithNewMessage(context.Background())
		req := new(Request)
		msg.WithClientReqHead(req)

		convey.Convey("GetGetInfluxDBClient fail: empty address", func() {
			opts := []transport.RoundTripOption{
				transport.WithDialAddress(emptyAddress),
			}
			_, err := tc.RoundTrip(ctx, nil, opts...)
			convey.So(err.Error(), convey.ShouldContainSubstring, "client transport get influxdb client error")
		})

		opts := []transport.RoundTripOption{
			transport.WithDialAddress(validAddress),
		}

		req.Op = 0
		convey.Convey("invalid Request Op", func() {
			_, err := tc.RoundTrip(ctx, nil, opts...)
			convey.So(err.Error(), convey.ShouldContainSubstring, "client transport invalid Request operation")
		})

		req.Op = OpPing
		req.Timeout = 10 * time.Second
		convey.Convey("ping fail", func() {
			_, err := tc.RoundTrip(ctx, nil, opts...)
			convey.So(err.Error(), convey.ShouldContainSubstring, "client transport: RspHead should be type of *Response")
		})

		msg.WithClientRspHead(new(PingResponse))
		convey.Convey("ping ok", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(tc), "getInfluxDBClient", func(_ *clientTransport, _ string) (*influxdbClient, error) {
				return &influxdbClient{Client: new(fakeClient)}, nil
			})
			defer patch.Reset()
			_, err := tc.RoundTrip(ctx, nil, opts...)
			t.Logf("err: %+v", err)
			convey.So(err, convey.ShouldBeNil)
		})

		req.Op = OpQuery
		msg.WithClientRspHead(nil)
		convey.Convey("ClientRspHead not *Response", func() {
			_, err := tc.RoundTrip(ctx, nil, opts...)
			convey.So(err.Error(), convey.ShouldContainSubstring, "client transport: RspHead should be type of *Response")
		})

		msg.WithClientRspHead(new(Response))
		convey.Convey("query fail", func() {
			_, err := tc.RoundTrip(ctx, nil, opts...)
			convey.So(err.Error(), convey.ShouldContainSubstring, "client transport influxdb Response error")
		})

		convey.Convey("query ok", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(tc), "getInfluxDBClient", func(_ *clientTransport, _ string) (*influxdbClient, error) {
				return &influxdbClient{Client: new(fakeClient)}, nil
			})
			defer patch.Reset()
			_, err := tc.RoundTrip(ctx, nil, opts...)
			convey.So(err, convey.ShouldBeNil)
		})

		req.Op = OpQueryAsChunk
		convey.Convey("invalid ChunkedResponse", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(tc), "getInfluxDBClient", func(_ *clientTransport, _ string) (*influxdbClient, error) {
				return &influxdbClient{Client: new(fakeClient)}, nil
			})
			defer patch.Reset()
			_, err := tc.RoundTrip(ctx, nil, opts...)
			convey.So(err.Error(), convey.ShouldContainSubstring, "client transport: RspHead should be type of *ChunkedResponse")
		})

		msg.WithClientRspHead(new(ChunkedResponse))

		convey.Convey("queryAsChunk ok", func() {
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(tc), "getInfluxDBClient", func(_ *clientTransport, _ string) (*influxdbClient, error) {
				return &influxdbClient{Client: new(fakeClient)}, nil
			})
			defer patch.Reset()
			_, err := tc.RoundTrip(ctx, nil, opts...)
			convey.So(err, convey.ShouldBeNil)
		})

		convey.Convey("write ok", func() {
			req.Op = OpWrite
			patch := gomonkey.ApplyPrivateMethod(reflect.TypeOf(tc), "getInfluxDBClient", func(_ *clientTransport, _ string) (*influxdbClient, error) {
				return &influxdbClient{Client: new(fakeClient)}, nil
			})
			defer patch.Reset()
			_, err := tc.RoundTrip(ctx, nil, opts...)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
