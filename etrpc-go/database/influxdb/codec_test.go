// Package influxdb provides influxdb v1 client
package influxdb

import (
	"context"
	"testing"

	"trpc.group/trpc-go/trpc-go/codec"

	"github.com/smartystreets/goconvey/convey"
)

var (
	cc  = &clientCodec{}
	msg = codec.Message(context.Background())
)

func TestClientCodec_Encode(t *testing.T) {
	convey.Convey("TestClientCodec_Encode", t, func() {
		_, err := cc.Encode(msg, nil)
		convey.So(err, convey.ShouldBeNil)
	})
}

func TestClientCodec_Decode(t *testing.T) {
	convey.Convey("TestClientCodec_Decode", t, func() {
		_, err := cc.Decode(msg, nil)
		convey.So(err, convey.ShouldBeNil)
	})
}
