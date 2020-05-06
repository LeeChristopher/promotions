package controllers

type ErrorController struct {
	BaseController
}

func (m *ErrorController) Error404() {
	m.SetResponse(404, "Page Not Found", nil)
}

func (m *ErrorController) Error500() {
	m.SetResponse(500, "Server Error", nil)
}
