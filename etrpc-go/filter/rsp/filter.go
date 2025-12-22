// Package rsp codec层的拦截器,对接口响应进行统一格式
package rsp

import (
	"context"
	"encoding/json"
	"etrpc-go/config"
	"etrpc-go/util/arrayutil"
	"etrpc-go/util/httputil"
	"net/http"
	"strconv"
	"strings"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	thttp "trpc.group/trpc-go/trpc-go/http"
)

// responseEntity 统一的服务端返回类型
type responseEntity struct {
	Code    int32           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	TraceId string          `json:"trace_id"`
}

// rspCfg rsp filter相关配置项
type rspCfg struct {
	DisableRspWrapper    bool     `yaml:"disable_rsp_wrapper"`     // 禁用响应体包装
	IgnoreRspWrapperPath []string `yaml:"ignore_rsp_wrapper_path"` // 忽略响应体包装
}

var cfg = &rspCfg{}

func init() {
	thttp.DefaultServerCodec.ErrHandler = errorHandler
	thttp.DefaultServerCodec.RspHandler = rspHandler
	config.RegisterConfigWithPrefix("etrpc.rsp_wrapper", "etrpc", cfg, true)
}

func errorHandler(w http.ResponseWriter, r *http.Request, e *errs.Error) {
	errMsg := strings.Replace(e.Msg, "\r", "\\r", -1)
	errMsg = strings.Replace(errMsg, "\n", "\\n", -1)

	// 保留trpc错误头信息,使用pb调用会用到
	w.Header().Add(thttp.TrpcErrorMessage, errMsg)
	if e.Type == errs.ErrorTypeFramework {
		w.Header().Add(thttp.TrpcFrameworkErrorCode, strconv.Itoa(int(e.Code)))
	} else {
		w.Header().Add(thttp.TrpcUserFuncErrorCode, strconv.Itoa(int(e.Code)))
	}
	// 映射code的标准http错误码
	if code, ok := thttp.ErrsToHTTPStatus[e.Code]; ok {
		w.WriteHeader(code)
	}

	// 判断是否需要对响应内容进行包装
	if !needWrap(r) {
		return
	}

	// 对错误信息进行封装统一返回
	resp := responseEntity{
		Code:    int32(e.Code),
		Message: errMsg,
		TraceId: getTraceId(r.Context()),
	}
	res, _ := json.Marshal(resp)
	_, _ = w.Write(res)
}

func rspHandler(w http.ResponseWriter, r *http.Request, body []byte) (err error) {
	// 只处理响应结构为json的请求,且body不为空,body为空代表标准的HTTP协议接口
	// 使用 application/json 前缀因为部分场景下会前端会使用 application/json;charset=UTF-8 调用
	if strings.HasPrefix(w.Header().Get(httputil.ContentTypeKey), httputil.ContentTypeJson) && body != nil {
		// 判断是否需要对响应内容进行包装
		if needWrap(r) {
			resp := responseEntity{
				Code:    0,
				Message: "success",
				TraceId: getTraceId(r.Context()),
				Data:    body,
			}
			body, err = json.Marshal(resp)
			if err != nil {
				return err
			}
		}
	}
	// 输出数据到前端
	_, err = w.Write(body)
	return err
}

// getTraceId 获取请求中的traceId
func getTraceId(ctx context.Context) string {
	msg := codec.Message(ctx)
	meta := msg.CommonMeta()
	if meta != nil {
		if traceId, ok := meta[httputil.EtrpcTraceIDKey]; ok {
			return traceId.(string)
		}
	}
	return ""
}

// needWrap 判断是否需要响应体包装
func needWrap(r *http.Request) bool {
	// 全局关闭wrapper
	if cfg.DisableRspWrapper {
		return false
	}
	// 请求Content-Type为PB,也不进行包装
	reqContentType := r.Header.Get(httputil.ContentTypeKey)
	if strings.EqualFold(reqContentType, httputil.ContentTypeProto) ||
		strings.EqualFold(reqContentType, httputil.ContentTypePB) {
		return false
	}
	// 部分接口关闭wrapper
	msg := codec.Message(r.Context())
	if arrayutil.Exist(cfg.IgnoreRspWrapperPath, msg.ServerRPCName()) {
		return false
	}
	// 调用方传metadata指定采用pb调用
	metaData := msg.ServerMetaData()
	if callType, ok := metaData[httputil.CallerTypeKey]; ok && string(callType) == httputil.CallerTypePB {
		return false
	}
	return true
}
