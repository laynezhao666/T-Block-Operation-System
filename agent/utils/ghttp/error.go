package ghttp

import (
	"fmt"
)

// Error represents a error.
type Error struct {
	Code    int
	Message string
}

// Error implements the error interface.
func (err Error) Error() string {
	return err.Message
}

// String implements the fmt.Stringer interface.
func (err Error) String() string {
	return fmt.Sprintf("code: %d, message: %s", err.Code, err.Message)
}

const (
	ErrorUnknown = 10001
)

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

// GetError returns a new Error from the given error.
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
