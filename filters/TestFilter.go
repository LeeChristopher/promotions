package filters

import "github.com/astaxie/beego/context"

type TestFilter struct {
	Request *context.BeegoInput
}

func NewTestFilter(request *context.BeegoInput) *TestFilter {
	return &TestFilter{Request: request}
}

func (m *TestFilter) IndexFilter() {

}
