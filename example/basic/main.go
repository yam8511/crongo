package main

import (
	"log"
	"time"

	"github.com/yam8511/crongo"
)

func main() {
	startTime := time.Now()
	// 程式離開前，最後一項任務
	defer func(startTime time.Time) {
		log.Println("Finish! Excursion:", time.Since(startTime))
	}(startTime)

	schdule := crongo.NewSchedule()
	one := schdule.NewShell(
		"zuolar",
		"*/2 * * * * *",
		"touch",
		[]string{time.Now().String()},
		false,
	)
	schdule.AddMission(one.Cron, one)
	schdule.Run()
	select {}
}
