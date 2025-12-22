// Package httputil uses for faster do http request
package httputil

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
	"trpc.group/trpc-go/trpc-go"

	"trpc.group/trpc-go/trpc-go/client"
	thttp "trpc.group/trpc-go/trpc-go/http"
)

// ResponseEntity Etrpc接口统一响应体,针对json响应体
type ResponseEntity[T any] struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	TraceId string `json:"trace_id"`
}

// ContentTypeKey 客户端请求头中Content-Type的值
const (
	ContentTypeKey   = "Content-Type"      // 客户端请求头中Content-Type的Key
	ContentTypeJson  = "application/json"  // 客户端请求头为Json的Content-Type值
	ContentTypeProto = "application/proto" // 客户端请求头为Proto的Content-Type值
	ContentTypePB    = "application/pb"    // 客户端请求头为Proto的Content-Type值
)

// 发送Http请求的一些默认参数
var (
	urlRegex = regexp.MustCompile(`^(http|https)://([^\s/]+)(/\S*)$`) // 匹配url的正则表达式
)

// 用于区分客户端调用方式,pb还是其他客户端,使用pb调用Etrpc服务必须通过httputil.GetPbCallOption()传递client.Option
const (
	CallerTypeKey = "call_type" // mateData中附带信息的key
	CallerTypePB  = "pb"        // mateData中附带信息的值

	EtrpcTraceIDKey = "etrpc_trace_id" // 用于向common metaData中设置trace_id的key
)

// GetPbCallOption
//
//	@Description:           用于使用桩代码调用Etrpc服务是附带传递的参数
//	@return client.Option	请求附带mateData信息的client.Option
func GetPbCallOption() client.Option {
	return client.WithMetaData(CallerTypeKey, []byte(CallerTypePB))
}

// GetJson
//
//	@Description:   使用http.Get发送json请求
//	@param ctx		上下文context
//	@param url		请求地址,支持ip+端口/域名/北极星服务名调用方式
//	@param header	额外的请求头信息,如鉴权信息,如果没有传递nil
//	@param resp		请求返回体,必须传递指针类型或者nil
//	@param opts		额外options，可以覆盖默认的options
//	@return error	请求错误信息
func GetJson(ctx context.Context, url string, header map[string]string, resp any, opts ...client.Option) error {
	if header == nil {
		header = map[string]string{}
	}
	header[ContentTypeKey] = ContentTypeJson
	_, err := Request(ctx, http.MethodGet, url, header, nil, resp, opts...)
	return err
}

// PostJson
//
//	@Description:   使用http.Post发送json请求
//	@param ctx		上下文context
//	@param url		请求地址,支持ip+端口/域名/北极星服务名调用方式
//	@param header	额外的请求头信息,如鉴权信息,如果没有传递nil
//	@param reqBody	请求参数体，必须传递指针类型或者nil
//	@param respBody	请求响应体, 必须传递指针类型或者nil
//	@param opts		额外options，可以覆盖默认的options
//	@return error	请求错误信息
func PostJson(ctx context.Context, url string, header map[string]string, reqBody, respBody any, opts ...client.Option) error {
	if header == nil {
		header = map[string]string{}
	}
	header[ContentTypeKey] = ContentTypeJson
	_, err := Request(ctx, http.MethodPost, url, header, reqBody, respBody, opts...)
	return err
}

// Get
//
//	@Description:		使用http.Get发送请求
//	@param ctx			上下文context
//	@param url			请求地址,支持ip+端口/域名/北极星服务名调用方式
//	@param header		额外的请求头信息,如鉴权信息,如果没有传递nil
//	@param resp			请求返回体,必须传递指针类型或者nil
//	@param opts			额外options，可以覆盖默认的options
//	@return error       请求错误信息
func Get(ctx context.Context, url string, header map[string]string, resp any, opts ...client.Option) error {
	_, err := Request(ctx, http.MethodGet, url, header, nil, resp, opts...)
	return err
}

// Post
//
//	@Description:		使用http.Post发送请求
//	@param ctx			上下文context
//	@param url			请求地址,支持ip+端口/域名/北极星服务名调用方式
//	@param header		额外的请求头信息,如鉴权信息,如果没有传递nil
//	@param req			请求参数体，必须传递指针类型或者nil
//	@param resp			请求响应体, 必须传递指针类型或者nil
//	@param opts			额外options，可以覆盖默认的options
//	@return error		请求错误信息
func Post(ctx context.Context, url string, header map[string]string, req, resp any, opts ...client.Option) error {
	_, err := Request(ctx, http.MethodPost, url, header, req, resp, opts...)
	return err
}

// Request
//
//	@Description:		发送http请求
//	@param context		上下文context
//	@param method		请求方式
//	@param url			请求地址,支持ip+端口/域名/北极星服务名调用方式
//	@param header		额外的请求头信息,如鉴权信息,如果没有传递nil
//	@param req			请求参数体，必须传递指针类型或者nil, Get会忽略掉这个参数
//	@param resp			请求响应体, 必须传递指针类型或者nil
//	@param opts			额外options，可以覆盖默认的options
//	@return rspHeader	请求响应头
//	@return err			请求错误信息
func Request(context context.Context, method string, reqUrl string, header map[string]string, req, resp any, opts ...client.Option) (rspHeader http.Header, err error) {
	method = strings.ToUpper(method)
	reqUrl = strings.TrimSpace(reqUrl)
	matches := urlRegex.FindStringSubmatch(reqUrl)
	if matches == nil {
		return rspHeader, fmt.Errorf("invalid reqUrl: %s, format must be [http|https]://[ip|ip:port|domain|domain:port|service-name]/xxx", reqUrl)
	}
	// 请求头
	reqHeader := &thttp.ClientReqHeader{}
	reqHeader.Method = method
	reqHeader.Schema = matches[1]
	if header != nil {
		for k, v := range header {
			reqHeader.AddHeader(k, v)
		}
	}
	// 返回头
	rspHead := &thttp.ClientRspHeader{}
	// 所有Options
	options := []client.Option{
		client.WithProtocol("http"),
		client.WithReqHead(reqHeader),
		client.WithRspHead(rspHead),
		client.WithTimeout(time.Second * 30),
	}
	service, path := matches[2], matches[3]
	// 非服务名的几种形式
	// 1、包含冒号,一般是ip/域名+端口号
	// 2、只有ip，默认80端口
	// 3、域名或者localhost，默认80端口
	// 其他的全部当做服务名处理
	if strings.Contains(service, ":") {
		split := strings.Split(service, ":")
		var target string
		if net.ParseIP(split[0]) != nil {
			target = fmt.Sprintf("ip://%s", service)
		} else {
			target = fmt.Sprintf("dns://%s", service)
		}
		options = append(options, client.WithTarget(target))
	} else if net.ParseIP(service) != nil {
		target := fmt.Sprintf("ip://%s", service)
		options = append(options, client.WithTarget(target))
	} else if addr, err := net.LookupHost(service); err == nil && len(addr) > 0 {
		target := fmt.Sprintf("dns://%s", service)
		options = append(options, client.WithTarget(target))
	} else {
		// 服务名类的
		options = append(options, client.WithServiceName(service))
		// 默认设置为全局的Namespace
		globalCfg := trpc.GlobalConfig().Global
		if globalCfg.Namespace != "" {
			options = append(options, client.WithNamespace(globalCfg.Namespace))
		}

		// 启用Set路由，则设置自动路由同Set服务
		if globalCfg.EnableSet == "Y" {
			// 判断是否主动禁用出站set路由
			optObj := &client.Options{}
			for _, opt := range opts {
				opt(optObj)
			}
			if !optObj.DisableServiceRouter {
				options = append(options, client.WithCallerSetName(globalCfg.FullSetName))
				options = append(options, client.WithCalleeSetName(globalCfg.FullSetName))
			}
		}
	}

	// 增加用户自定义options
	options = append(options, opts...)
	// 创建http客户端
	httpCli := thttp.NewClientProxy(service, options...)
	// 发起请求
	switch method {
	case http.MethodGet:
		err = httpCli.Get(context, path, resp)
	case http.MethodPost:
		err = httpCli.Post(context, path, req, resp)
	case http.MethodPut:
		err = httpCli.Put(context, path, req, resp)
	case http.MethodDelete:
		err = httpCli.Delete(context, path, req, resp)
	case http.MethodPatch:
		err = httpCli.Patch(context, path, req, resp)
	default:
		return rspHeader, errors.New("invalid http method")
	}
	if err != nil {
		return rspHeader, err
	}
	return rspHead.Response.Header, nil
}
