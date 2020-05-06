package controllers

import (
	"promotions/filters"
	"promotions/packages/tools"
)

var (
	authFilter *filters.AuthFilter
)

type AuthController struct {
	BaseController
}

func (m *AuthController) Initialise() {
	authFilter = filters.NewAuthFilter(m.Ctx.Request)
}

func (m *AuthController) Login() {
	info, err := authFilter.Login()
	if err != nil {
		m.SetResponse(tools.CodeMap["fail"], err.Error(), nil)
		return
	}

	m.SetResponse(tools.CodeMap["success"], "", info)
}
