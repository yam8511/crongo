package crongo

import (
	"log"
	"os"
)

// DebugMode Debug模式
var DebugMode bool

// 如同 nodejs 的 indexOf
// 從 arr 裡面找出 val 的位子
func indexOf(arr []int, val int) int {
	for index, value := range arr {
		if value == val {
			return index
		}
	}
	return -1
}

func writeLog(message ...interface{}) {
	if !DebugMode && os.Getenv("GOCRON_MODE") == "release" {
		return
	}

	log.Println(message...)
}
