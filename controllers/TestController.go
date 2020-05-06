package controllers

import (
	"promotions/filters"
	"promotions/models/users"
	"promotions/packages/tools"
)

var testFilter *filters.TestFilter

type TestController struct {
	BaseController
}

func (m *TestController) Initialise() {
	testFilter = filters.NewTestFilter(m.Ctx.Input)
}

func (m *TestController) Index() {
	testFilter.IndexFilter()
	m.SetResponse(tools.CodeMap["success"], "ok", users.LoginUserInfo.Username)
}
