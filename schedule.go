package crongo

import (
	cron "gopkg.in/robfig/cron.v2"
)

// Schedule : 背景排程
type Schedule struct {
	// Missions : 需要執行的背景任務
	Missions []cron.Job
	Running  bool
	Cron     *cron.Cron
}

// NewSchedule : 建立一個新排程
func NewSchedule() *Schedule {
	newSchedule := &Schedule{
		Missions: []cron.Job{},
		Running:  false,
		Cron:     cron.New(),
	}
	return newSchedule
}

// NewShell : 建立一個新任務
func (schedule *Schedule) NewShell(name, cron, command string, args []string, overlapping bool) *Shell {
	return &Shell{
		Name:        name,
		Cron:        cron,
		Command:     command,
		Args:        args,
		Overlapping: overlapping,
		Pids:        []int{},
	}
}

// AddMission : 新增任務到背景排程
func (schedule *Schedule) AddMission(cron string, mission cron.Job) error {
	schedule.Missions = append(schedule.Missions, mission)
	_, err := schedule.Cron.AddJob(cron, mission)
	return err
}

// Run : 開始執行背景排程
func (schedule *Schedule) Run() {
	// 如果背景正在跑，則跳過
	if schedule.Running {
		return
	}

	schedule.Running = true
	schedule.Cron.Start()
}

// Suspend : 停止背景排程
func (schedule *Schedule) Suspend() {
	if !schedule.Running {
		return
	}
	schedule.Running = false
	schedule.Cron.Stop()
}
