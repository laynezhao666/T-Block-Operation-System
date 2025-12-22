package usr

import (
	"agent/entity/errcode"

	"trpc.group/trpc-go/trpc-go/errs"
)

// Login 登录
func Login(userName string, password string) (string, error) {
	id, err := Matched(userName, password)
	if err != nil {
		return "", err
	}

	token, err := Signature(
		Context{
			ID:       id,
			UserName: userName,
		}, "",
	)
	if err != nil {
		return "", errs.New(errcode.ErrCgiUserSignatureFail, "login signature fail")
	}
	return token, nil
}

// Matched todo 临时方案，仅提供admin默认用户
func Matched(userName string, password string) (uint, error) {
	if userName == "admin" && password == "tencent@123" {
		return 1, nil
	}
	return 0, errs.New(errcode.ErrCgiUserLoginFail, "username or password invalid")
}
