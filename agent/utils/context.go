package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"trpc.group/trpc-go/trpc-go/codec"
	thttp "trpc.group/trpc-go/trpc-go/http"
)

// IsContextDone 判断上下文是否结束
func IsContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// HeaderMatcher 目的是获取到请求的Body，因为restful在进入实际的handle函数之前，框架会读取Body进行转码，映射到对应的Pb的Message上面
func HeaderMatcher(ctx context.Context, w http.ResponseWriter, r *http.Request, serviceName, methodName string,
) (context.Context, error) {
	ctx, msg := codec.WithNewMessage(ctx)
	msg.WithCalleeServiceName(serviceName)
	msg.WithServerRPCName(methodName)
	msg.WithSerializationType(codec.SerializationTypePB)
	// 然后将 req.Body 读取完之后重新进行设置
	buf, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(buf))
	header := thttp.Head(ctx)
	if header == nil {
		return nil, fmt.Errorf("获取不到 ctx 中的 header")
	}
	header.ReqBody = buf
	return ctx, nil
}
