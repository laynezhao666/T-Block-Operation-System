package utils

import (
	"context"
	"strings"

	"etrpc-go/log"

	"trpc.group/trpc-go/trpc-go/codec"
)

// GetUpstreamIp 通过ctx获取请求上游的IP
func GetUpstreamIp(ctx context.Context) string {
	c := ctx.Value(codec.ContextKeyMessage)
	val, ok := c.(codec.Msg)
	if !ok {
		log.WarnContextf(ctx, "Get Upstream IP err: cannot convert to codec.Msg")
		return ""
	}
	// addr.String()格式为[ip:port]如127.0.0.1:1234
	addr := val.RemoteAddr()
	idx := strings.Index(addr.String(), ":")
	if idx == -1 {
		log.WarnContextf(ctx, "Get Upstream IP err: remote addr <%v> has no \":\"", addr)
		return ""
	}
	return addr.String()[:idx]
}
