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
		[]string{"1"},
		[]string{},
		false,
		false,
		true,
		nil,
		nil,
		nil,
	)
	twice := schdule.NewShell(
		"snoopy",
		"* * * * * *",
		"sleep",
		[]string{"3"},
		[]string{},
		true,
		false,
		true,
		nil,
		nil,
		nil,
	)
	schdule.AddMission(one)
	schdule.AddMission(twice)
	schdule.Start()
	time.Sleep(time.Second * 5)
	schdule.Stop()
}
