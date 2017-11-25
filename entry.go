package main

import (
	"log"

	cron "gopkg.in/robfig/cron.v2"
)

// Schedule : 背景排程
type Schedule struct {
	// Missions : 需要執行的背景任務
	Missions []Mission
}

// Mission : 正在執行的背景任務
type Mission struct {
	// 任務名稱
	name string
	// 執行週期
	cron string
	// 指令
	command string
	// 指令參數
	args []string
	// 是否能重複行
	overlapping bool
	// 已執行的PID
	pid []int
}

// Run : 開始執行背景
func (schedule *Schedule) Run() {
	i := 0
	c := cron.New()
	spec := "*/5 * * * * ?"
	c.AddFunc(spec, func() {
		log.Println("En", c.Entries())
		i++
		log.Println("cron running:", i)
	})
	c.Start()

	select {}
}
