// Package ghttp 提供门禁CGI服务的HTTP响应封装和错误码定义。
package ghttp

import (
	"fmt"
)

// Error HTTP业务错误类型，包含错误码和错误信息
type Error struct {
	Code    int    // 错误码
	Message string // 错误信息
}

// Error 实现error接口
func (err Error) Error() string {
	return err.Message
}

// String 返回错误的格式化字符串
func (err Error) String() string {
	return fmt.Sprintf("code: %d, message: %s", err.Code, err.Message)
}

// ErrorUnknown 未知错误码
const (
	ErrorUnknown = 10001
)

// ErrorOK 成功响应
// ErrorInternalServer 内部服务错误
// ErrorBind 请求体绑定错误
// ErrorDatabase 数据库错误
// ErrorCreate 创建操作错误
// ErrorNotFound 资源未找到
// ErrorDelete 删除操作错误
// ErrorUpdate 更新操作错误
// ErrorRead 读取操作错误
// ErrorParamInvalid 参数无效
// ErrorParse 请求体解析错误
// ErrorUploadFileInvalid 上传文件无效
// ErrorUploadFileError 上传文件格式错误
// ErrorModifyFileError 修改文件错误
// ErrorControlError 控制命令错误
// ErrorCGIReqError CGI请求发送错误
// ErrorCGIRespError CGI响应解析错误
// ErrorCGITopicError CGI主题错误
// ErrorApiReqError API请求发送错误
// ErrorTokenInvalid Token无效
// ErrorTokenGenerate Token生成错误
var (
	ErrorOK = &Error{Code: 0, Message: "OK"}

	ErrorInternalServer = &Error{Code: 10001, Message: "Internal server error when cast error type to errno.Error type"}
	ErrorBind           = &Error{Code: 10002, Message: "Bind the post request body to struct error"}

	ErrorDatabase = &Error{Code: 20001, Message: "Database error"}
	ErrorCreate   = &Error{Code: 20002, Message: "Create error"}
	ErrorNotFound = &Error{Code: 20003, Message: "Not found"}
	ErrorDelete   = &Error{Code: 20004, Message: "Delete error"}
	ErrorUpdate   = &Error{Code: 20005, Message: "Update error"}
	ErrorRead     = &Error{Code: 20006, Message: "Read error"}

	ErrorParamInvalid      = &Error{Code: 20101, Message: "Params not valid"}
	ErrorParse             = &Error{Code: 20102, Message: "Parse post body error"}
	ErrorUploadFileInvalid = &Error{Code: 20103, Message: "File to upload not valid"}
	ErrorUploadFileError   = &Error{Code: 20104, Message: "File format error or path not valid"}
	ErrorModifyFileError   = &Error{Code: 20105, Message: "Modify file error"}
	ErrorControlError      = &Error{Code: 20106, Message: "Control command error"}

	ErrorCGIReqError   = &Error{Code: 30101, Message: "Send req to tbmon CGI error"}
	ErrorCGIRespError  = &Error{Code: 30102, Message: "Parse resp from tbmon CGI error"}
	ErrorCGITopicError = &Error{Code: 30103, Message: "topic of CGI error"}

	ErrorApiReqError = &Error{Code: 30201, Message: "Send req to tbcm api error"}

	ErrorTokenInvalid  = &Error{Code: 40101, Message: "token is invalid"}
	ErrorTokenGenerate = &Error{Code: 40102, Message: "generate token error"}
)

// GetError 将标准error转换为业务Error类型
func GetError(err error) *Error {
	if err == nil {
		return ErrorOK
	}
	switch e := err.(type) {
	case *Error:
		return e
	}
	return &Error{
		Code:    ErrorUnknown,
		Message: fmt.Sprintf("%+v", err),
	}
}
