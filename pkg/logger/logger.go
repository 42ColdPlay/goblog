package logger

import "log"

//LogError当存在错误时记录日志
func LogError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
