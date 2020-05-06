package tools

var CodeMap map[string]int

func InitCode() {
	CodeMap = map[string]int{
		"success":      200,
		"fail":         0,
		"serverError":  500,
		"pageNotFound": 404,
	}
}
