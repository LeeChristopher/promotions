package filters

import (
	"net/http"
	"promotions/models/users"
	"promotions/services"
	"regexp"

	"github.com/astaxie/beego/validation"
)

type AuthFilter struct {
	Request *http.Request
}

func NewAuthFilter(request *http.Request) *AuthFilter {
	return &AuthFilter{Request: request}
}

func (m *AuthFilter) Login() (loginResult users.LoginUserResponse, err error) {
	username := m.Request.FormValue("username")
	password := m.Request.FormValue("password")

	valid := validation.Validation{}
	valid.Required(username, "username").Message("请填写用户名！")
	valid.Match(username, regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`), "username").Message("用户名仅支持英文、数字且以英文开头！")
	valid.MinSize(username, 5, "password").Message("用户名最少填写5位")
	valid.MaxSize(username, 30, "username").Message("用户名最大长度支持30位")
	valid.Required(password, "password").Message("请填写密码！")
	valid.Match(password, regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`), "password").Message("密码仅支持英文、数字且以英文开头！")
	valid.MinSize(password, 8, "password").Message("密码最少填写8位")
	valid.MaxSize(password, 30, "password").Message("密码最大长度支持30位")
	if valid.HasErrors() {
		return users.LoginUserResponse{}, valid.Errors[0]
	}

	loginResult, err = services.Login(username, password)
	if err != nil {
		return loginResult, err
	}

	return loginResult, nil
}
