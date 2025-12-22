package logic

import (
	"context"
	"reflect"
	"strconv"

	"trpc.group/trpc-go/trpc-go/filter"
	thttp "trpc.group/trpc-go/trpc-go/http"
)

// MozuIdServerFilter 服务端模组ID拦截器,将请求头中的MozuId取出来放在请求体中
func MozuIdServerFilter(ctx context.Context, req interface{}, next filter.ServerHandleFunc) (rsp interface{}, err error) {
	if head := thttp.Head(ctx); head != nil {
		// 取请求头中的mozuid字段
		mozuIdStr := head.Request.Header.Get("mozuid")
		if mozuId, err := strconv.Atoi(mozuIdStr); err == nil {
			if field := reflect.ValueOf(req).Elem().FieldByName("MozuId"); field.IsValid() {
				if field.Kind() == reflect.Int32 || field.Kind() == reflect.Int64 || field.Kind() == reflect.Int {
					// 仅当未设置值的时候设置
					if field.IsZero() {
						field.SetInt(int64(mozuId))
					}
				} else if field.Kind() == reflect.String {
					// 仅当未设置值的时候设置
					if field.IsZero() {
						field.SetString(mozuIdStr)
					}
				}
			}
		}
	}
	return next(ctx, req)
}
