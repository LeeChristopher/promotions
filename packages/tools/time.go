package tools

import "time"

const DATA_FORMART = "2006-01-02"
const DATATIME_FORMART = "2006-01-02 15:04:05"
const LOCATION = "Asia/Shanghai"

func GetNow() time.Time {
	location, _ := time.LoadLocation(LOCATION)
	return time.Now().In(location)
}
