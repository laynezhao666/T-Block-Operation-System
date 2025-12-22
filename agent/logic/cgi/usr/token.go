package usr

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const (
	LoginStatusExpireTime = 6 * 3600
)

var (
	ErrAuthorizationNotInRequestHeader = errors.New("not find Authorization in request header")
	ErrTokenNotValid                   = errors.New("token is not valid")
)

func secretFunc(secret string) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	}
}
// ParseToken 解析token
func ParseToken(tokenStr, secret string) (*Context, error) {
	token, err := jwt.Parse(tokenStr, secretFunc(secret))
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &Context{
			ID:       uint(claims["id"].(float64)),
			UserName: claims["username"].(string),
		}, nil
	} else {
		if !ok {
			return nil, fmt.Errorf("convert %+v to jwt.MapClaims failed", token.Claims)
		}
		return nil, ErrTokenNotValid
	}
}

func getSecret() string {
	return "123todo"
}
// ParseRequest 解析请求
func ParseRequest(c *gin.Context) (*Context, error) {
	//TODO: 携带 token 时才进行校验?
	header := c.Request.Header.Get("Authorization")
	if len(header) == 0 {
		return nil, nil
		//return nil, ErrAuthorizationNotInRequestHeader
	}
	token := ""
	_, err := fmt.Sscanf(header, "Bearer %s", &token)
	if err != nil {
		return nil, err
	}
	secret := getSecret()
	return ParseToken(token, secret)
}
// Signature 签名
func Signature(ctx Context, secret string) (tokenStr string, err error) {
	if secret == "" {
		secret = getSecret()
	}
	currentTime := time.Now().Unix()
	expireTime := currentTime + int64(LoginStatusExpireTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       ctx.ID,
		"username": ctx.UserName,
		"nbf":      currentTime,
		"iat":      currentTime,
		"exp":      expireTime,
	})
	return token.SignedString([]byte(secret))
}
