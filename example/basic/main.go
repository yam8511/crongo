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
		"* * * * * *",
		"sleep",
		[]string{"3"},
		false,
		true,
	)
	twice := schdule.NewShell(
		"snoopy",
		"* * * * * *",
		"sleep",
		[]string{"3"},
		true,
		true,
	)
	schdule.AddMission(one.Cron, one)
	schdule.AddMission(twice.Cron, twice)
	schdule.Run()
	time.Sleep(time.Second * 10)
	schdule.Suspend()
}
