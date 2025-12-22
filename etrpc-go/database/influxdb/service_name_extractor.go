// Package influxdb provides influxdb v1 client
package influxdb

import (
	"errors"
	"strings"
)

// URIHostExtractor 从URI中提取host，用于ip解析（如host再从北极星查询ip），配合ResolvableSelector使用
type URIHostExtractor struct {
}

// Extract 这里的uri已经剔除了“://”及它之前的部分
func (e *URIHostExtractor) Extract(uri string) (int, int, error) {
	// influxdb+polaris://user:password@trpc.influxdb.xxx.xxx?timeout=1
	offset := 0

	// host起始位置部分
	if idx := strings.LastIndex(uri, "@"); idx != -1 {
		uri = uri[idx+1:]
		offset += idx + 1
	}

	// 解析host结束位置部分
	begin := offset
	length := len(uri)
	if idx := strings.IndexAny(uri, "/?@"); idx != -1 {
		if uri[idx] == '@' {
			return 0, 0, errors.New("parse host from uri: unescaped @ sign in user info")
		}
		length = idx
	}
	uri = uri[0:length]

	// 跳过协议字符tcp( 或者 )等token
	return e.dealProtocolToken(uri, begin, length)
}

func (e *URIHostExtractor) dealProtocolToken(uri string, begin, length int) (int, int, error) {
	if strings.HasPrefix(uri, "tcp(") {
		begin += 4
		length -= 4
	}
	if strings.HasSuffix(uri, ")") {
		length--
	}
	return begin, length, nil
}
