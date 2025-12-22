package errcode

const (
	// 通用错误码
	ErrCgiParamInvalid = iota + 270200
	ErrCgiHttpMethodNotSupported
	ErrNilHeader
	ErrCgiNotImplemented
	ErrCgiHandleFail
	// 具体错误码
	ErrCgiDeviceIdEmpty
	ErrCgiDeviceValueDefErr
	ErrCgiTemplateFileFail
	ErrSaveConfigFail
	ErrCgiUserLoginFail
	ErrCgiUserSignatureFail
	ErrCgiGetMappingFail
	ErrCgiSaveMappingFail
)

const (
	DefaultCgiRspCode         = 0
	DefaultCgiRspMessage      = "success"
	ErrCgiTemplateFileMessage = "表格导入失败"
)
