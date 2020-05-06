package tools

import (
	"encoding/json"
	"net/http"
	"time"
)

type MiddlewareHandler struct {
	httpHandler   http.Handler
	requestMethod map[string]uint8
}

/**
初始化中间件结构体
*/
func NewMiddlewareHandler(httpHandler http.Handler) *MiddlewareHandler {
	return &MiddlewareHandler{
		httpHandler: httpHandler,
		requestMethod: map[string]uint8{
			"GET":  1,
			"POST": 1,
		},
	}
}

/**
继承上级接口
*/
func (m *MiddlewareHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	startTime := GetNow()
	//设置跨域
	m.setCross(response)

	//请求方法设置
	if _, ok := m.requestMethod[request.Method]; !ok {
		commonResponse := CommonResponse{
			RunTime: time.Since(startTime).Seconds(),
			Code:    CodeMap["fail"],
			Message: "请求方式错误！",
			Data:    nil,
		}
		commonResponseByte, _ := json.Marshal(&commonResponse)
		_, _ = response.Write(commonResponseByte)
		return
	}

	//传递给下层
	_ = request.ParseForm()
	m.httpHandler.ServeHTTP(response, request)
}

/**
设置跨域
*/
func (m *MiddlewareHandler) setCross(response http.ResponseWriter) {
	response.Header().Set("Access-Control-Allow-Origin", "*")
	response.Header().Set("Access-Control-Allow-Headers", "Content-Type, MHToken")
	response.Header().Set("content-type", "application/json")
}
