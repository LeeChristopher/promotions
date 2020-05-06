package tools

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

type CommonResponse struct {
	RunTime float64     `json:"runtime"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type UserInfo struct {
	UserId   uint64
	Username string
	Email    string
}

type AuthToken struct {
	UserInfo
	jwt.StandardClaims
}

func IssueAuthToken(userInfo UserInfo) (result bool, tokenString string) {
	claim := AuthToken{
		UserInfo: userInfo,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: GetNow().Add(time.Minute).Unix(),
			IssuedAt:  GetNow().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	tokenString, err := token.SignedString([]byte(AppConfig.SecretKey))
	if err != nil {
		return false, ""
	}

	return true, tokenString
}
