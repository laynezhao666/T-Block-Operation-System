// Package influxdb provides influxdb v1 client
package influxdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractHost(t *testing.T) {

	testCases := []struct {
		uri  string // 这里的uri已经剔除了“://”及它之前的部分
		host string
		err  string
	}{
		{
			uri:  "localhost",
			host: "localhost",
		},
		{
			uri:  "admin:123456@localhost/",
			host: "localhost",
		},
		{
			uri:  "admin:123456@localhost",
			host: "localhost",
		},
		{
			uri:  "example1.com:27017,example2.com:27017",
			host: "example1.com:27017,example2.com:27017",
		},
		{
			uri:  "host1,host2,host3/?slaveOk=true",
			host: "host1,host2,host3",
		},
		{
			uri:  "admin:123456@localhost?",
			host: "localhost",
		},
		{
			uri:  "user:secret@localhost:6379/0?foo=bar&qux=baz", // redis示例
			host: "localhost:6379",
		},
		{
			uri:  "user:secretWith@secretWith@localhost:6379/0?foo=bar&qux=baz", // 密码包含@符号示例
			host: "localhost:6379",
		},
		{
			uri:  "user:secret@tcp(localhost:6379)/Database?timeout=1s&interpolateParams=true", // mysql示例
			host: "localhost:6379",
		},
	}

	extractor := new(URIHostExtractor)
	for _, tc := range testCases {
		t.Run(tc.uri, func(t *testing.T) {
			pos, length, err := extractor.Extract(tc.uri)
			if len(tc.err) != 0 {
				assert.EqualError(t, err, tc.err)
			} else {
				assert.Equal(t, tc.host, tc.uri[pos:pos+length])
			}
		})
	}
}
