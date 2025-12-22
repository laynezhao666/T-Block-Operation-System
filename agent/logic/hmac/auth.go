package hmac

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"agent/entity/config"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"trpc.group/trpc-go/trpc-go/errs"
	"trpc.group/trpc-go/trpc-go/filter"
	thttp "trpc.group/trpc-go/trpc-go/http"
	"trpc.group/trpc-go/trpc-go/log"
)

const (
	HeaderOfAuth   = "Authorization"
	HMACSHA256     = "HMAC-SHA-256"
	maxReqInterval = 600
)

// AuthInfo 鉴权信息
type AuthInfo struct {
	Alg       string
	SystemID  string
	Timestamp string
	Signature string
}

// 初始化
func init() {
	// 示例配置，实际应从配置中心获取
	//authConfig = &config.AuthConf{
	//	SysID2Key: map[string]string{
	//		"101": "f9b3d2e7a4c5816b0e5f8d3c7a9b4e1",
	//	},
	//}
}

// NewAuthFilter 创建鉴权过滤器
func NewAuthFilter() filter.ServerFilter {
	return func(ctx context.Context, req interface{}, next filter.ServerHandleFunc) (interface{}, error) {
		head := thttp.Head(ctx)
		if head == nil {
			return nil, errs.New(400, "BadRequest, 请求头解析失败")
		}

		// 获取当前请求路径
		currentPath := head.Request.URL.Path

		// 检查是否需要鉴权
		if !needAuth(currentPath) {
			// 不需要鉴权的接口直接放行
			return next(ctx, req)
		}

		// 解析认证信息
		authStr := head.Request.Header.Get(HeaderOfAuth)
		info := parseAuthStr(authStr)
		if info == nil || info.IsEmpty() {
			return nil, errors.New("invalid hmac header")
		}

		// 获取请求体
		body, err := getRequestBody(ctx)
		if err != nil {
			return nil, fmt.Errorf("read body error: %v", err)
		}

		// 执行鉴权验证
		if err := verifyAuth(info, body, head.Request.Header); err != nil {
			return nil, err
		}

		return next(ctx, req)
	}
}

// 路由匹配逻辑
func needAuth(currentPath string) bool {
	authConf := config.GetRB().AuthConf

	// 总开关关闭时不需要鉴权
	if !authConf.SecurityConf.Enabled {
		return false
	}

	// 遍历配置的鉴权路由
	for _, pattern := range authConf.RequiredRoutes {
		// 使用path匹配支持通配符
		if pathMatch(pattern, currentPath) {
			return true
		}
	}
	return false
}

// 路径匹配，支持通配符*
func pathMatch(pattern, target string) bool {
	if pattern == "*" || pattern == "/*" {
		return true
	}

	// 处理结尾通配符
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(target, prefix)
	}

	// 精确匹配
	return path.Clean(pattern) == path.Clean(target)
}

func parseAuthStr(authStr string) *AuthInfo {
	if authStr == "" {
		return nil
	}

	parts := strings.SplitN(authStr, " ", 2)
	if len(parts) != 2 {
		return nil
	}

	info := &AuthInfo{Alg: parts[0]}
	params := strings.Split(parts[1], ",")

	for _, param := range params {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) != 2 {
			continue
		}

		switch kv[0] {
		case "SystemId":
			info.SystemID = kv[1]
		case "Timestamp":
			info.Timestamp = kv[1]
		case "Signature":
			info.Signature = kv[1]
		}
	}
	return info
}

// IsEmpty 判断是否为空
func (a *AuthInfo) IsEmpty() bool {
	return a.Alg == "" || a.SystemID == "" || a.Timestamp == "" || a.Signature == ""
}

func getRequestBody(ctx context.Context) (string, error) {
	head := thttp.Head(ctx)
	if head == nil {
		return "", errors.New("invalid request head")
	}
	body := head.ReqBody
	return string(body), nil

	//return "", errors.New("invalid request body type")
}

func verifyAuth(info *AuthInfo, body string, header http.Header) error {
	// 时间戳格式校验
	reqTs, err := strconv.ParseInt(info.Timestamp, 10, 64)
	if err != nil {
		return errs.New(400, "BadRequest, 时间戳格式错误")
	}

	// 请求时效性校验
	if time.Now().Unix() > reqTs+maxReqInterval {
		return errs.New(400, "BadRequest, 请求已过期")
	}

	// 系统ID有效性校验
	secretKey := config.GetRB().GetSysKeyById(info.SystemID)
	log.Debugf("SystemID:%s,secretKey:%s", info.SystemID, secretKey)
	if secretKey == "" {
		return errs.New(403, "Forbidden, 系统ID无效或过期")
	}

	// 签名生成与比对
	signData := fmt.Sprintf("%d%s", reqTs, body)
	expectedSign := generateHMAC(signData, secretKey, info.Alg)
	if !hmac.Equal([]byte(expectedSign), []byte(info.Signature)) {
		return errs.New(401, "Unauthorized, 签名验证失败")
	}
	return nil
}

func generateHMAC(data, key, alg string) string {
	var hashFunc func() hash.Hash

	switch alg {
	case HMACSHA256:
		hashFunc = sha256.New
	default:
		hashFunc = sha1.New
	}
	// 约定提供的key为十六进制字符串，需要转换为字节数组
	keyByte, err := hex.DecodeString(key)
	if err != nil {
		return ""
	}
	mac := hmac.New(hashFunc, keyByte)
	//mac := hmac.New(hashFunc, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
