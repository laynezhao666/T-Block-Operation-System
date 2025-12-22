// Package userutil provides utility functions for get current login user
package userutil

import "context"

// User 当前用户信息
type User struct {
	StaffId  string `json:"staff_id"`
	Username string `json:"username"`
}

// GetLoginUser 获取当前请求登录的用户信息
//
//	@param ctx		context.Context
//	@return *User	当前登录的用户信息，如果未登录则为nil
func GetLoginUser(ctx context.Context) *User {
	// TODO: 从ctx中获取用户信息
	return &User{}
}

// GetLoginUsername 获取当前请求登录的用户名
//
//	@param ctx			context.Context
//	@return string		当前登录的用户名，如果未登录则为空字符串
func GetLoginUsername(ctx context.Context) string {
	user := GetLoginUser(ctx)
	if user == nil {
		return ""
	}
	return user.Username
}
