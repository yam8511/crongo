package crongo

import (
	"log"

	cron "gopkg.in/robfig/cron.v2"
)

// Schedule : 背景排程
type Schedule struct {
	// Missions : 需要執行的背景任務
	Missions []Mission
	Running  bool
	Cron     *cron.Cron
}

// Mission : 任務介面
type Mission interface {
	Run()
	Disable()
	GetName() string
	GetPids() []int
}

// NewSchedule : 建立一個新排程
func NewSchedule() *Schedule {
	newSchedule := &Schedule{
		Missions: []Mission{},
		Running:  false,
		Cron:     cron.New(),
	}
	return newSchedule
}

// NewShell : 建立一個新任務
func (schedule *Schedule) NewShell(name, cron, command string, args []string, overlapping, enable bool) *Shell {
	return &Shell{
		Name:        name,
		Cron:        cron,
		Command:     command,
		Args:        args,
		Overlapping: overlapping,
		Pids:        []int{},
		Enable:      enable,
	}
}

// AddMission : 新增任務到背景排程
func (schedule *Schedule) AddMission(cron string, mission Mission) error {
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

	log.Println("====== 等待背景以下程序結束... ======")
	for _, mission := range schedule.Missions {
		mission.Disable()
		if pids := mission.GetPids(); len(pids) > 0 {
			log.Printf("> Command:〈 %s 〉, PID:〈 %d 〉\n", mission.GetName(), pids)
		}
	}
	log.Println("====================================")

	waittingProcessFinish := make(chan int)
	go func() {
		for {
			hasPID := false
			for _, mission := range schedule.Missions {
				if len(mission.GetPids()) > 0 {
					hasPID = true
				}
			}

			if !hasPID {
				break
			}
		}
		waittingProcessFinish <- 0
	}()

	<-waittingProcessFinish

	schedule.Running = false
	schedule.Cron.Stop()
}
