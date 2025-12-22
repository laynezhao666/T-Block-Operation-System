// Package influxdb provides influxdb v1 client
package influxdb

import (
	"fmt"
	"os"
	"path"

	"trpc.group/trpc-go/trpc-go/codec"
)

func init() {
	codec.Register(databaseName, nil, &clientCodec{})
}

// ClientCodec client codec
type clientCodec struct{}

// Encode set meta data
func (c *clientCodec) Encode(msg codec.Msg, body []byte) (buffer []byte, err error) {
	if msg.CallerServiceName() == "" {
		msg.WithCallerServiceName(fmt.Sprintf("trpc.%s.%s.service", databaseName, path.Base(os.Args[0])))
	}
	return nil, nil
}

// Decode nil decode
func (c *clientCodec) Decode(message codec.Msg, buffer []byte) (body []byte, err error) {
	return nil, nil
}
