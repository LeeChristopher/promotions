package tools

import "strings"

var noSignUrl = map[string]bool{
	"/login":     true,
	"/register":  true,
	"/promotion": true,
}

func GetIsSign(url string) bool {
	if url == "" {
		return false
	}
	if strings.Contains(url, "?") {
		index := strings.Index(url, "?")
		url = url[0:index]
	}

	if _, ok := noSignUrl[url]; !ok {
		return false
	}

	return true
}
