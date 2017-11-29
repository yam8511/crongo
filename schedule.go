package crongo

import (
	"fmt"
	"log"
	"os/exec"
	"time"

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
	Enable()
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
func (schedule *Schedule) NewShell(name, cron, command string, args []string, overlapping, enable bool, errorHandler func(error)) *Shell {
	handler := errorHandler
	if errorHandler == nil {
		handler = func(err error) {
			log.Printf("[Warning] Command:〈 %s 〉throw error %v\n", name, err)
		}
	}
	return &Shell{
		Name:         name,
		Cron:         cron,
		Command:      command,
		Args:         args,
		Overlapping:  overlapping,
		Pids:         []int{},
		IsEnable:     enable,
		ErrorHandler: handler,
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
		log.Println("======= 目前背景排程已經啟動... =======")
		return
	}

	log.Println("====== !!! 開始啟動背景程序 !!! ======")
	schedule.Running = true
	schedule.Cron.Start()
}

// Suspend : 停止背景排程
func (schedule *Schedule) Suspend() {
	if !schedule.Running {
		log.Println("======= 目前背景排程已經停止... =======")
		return
	}

	schedule.Cron.Stop()

	log.Println("======= 等待背景以下程序結束... =======")
	for _, mission := range schedule.Missions {
		if pids := mission.GetPids(); len(pids) > 0 {
			log.Printf("> Command:〈 %s 〉, PID:〈 %d 〉\n", mission.GetName(), pids)
		}
	}
	log.Println("=======================================")

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
	log.Println("======== !!! 背景程序已暫停 !!! =======")
}

// Destroy : 強制停止背景排程
func (schedule *Schedule) Destroy() {
	if !schedule.Running {
		log.Println("======== 目前背景排程尚未啟動... ========")
		return
	}
	schedule.Cron.Stop()

	log.Println("======== !!! 即將摧毀以下背景程序 !!! ========")
	for _, mission := range schedule.Missions {
		if pids := mission.GetPids(); len(pids) > 0 {
			log.Printf("> Command:〈 %s 〉, PID:〈 %d 〉\n", mission.GetName(), pids)
		}
	}
	log.Println("==============================================")

	timer := time.NewTicker(time.Second)
	tick := 5
	go func() {
		for range timer.C {
			log.Printf("[Danger] 倒數%d秒.....\n", tick)
			tick--
		}
	}()
	for tick > 0 {
	}
	timer.Stop()

	log.Println("========== !!! 開始摧毀背景程序 !!! ==========")
	killPids := []string{}
	for _, mission := range schedule.Missions {
		for _, pid := range mission.GetPids() {
			killPids = append(killPids, fmt.Sprint(pid))
		}
	}
	killer := exec.Command("kill", killPids...)
	err := killer.Run()
	time.Sleep(time.Second)
	if err != nil {
		log.Println("========== !!! 背景摧毀發生錯誤 !!! ==========")
		log.Printf("[Error] %v\n", err)
	} else {
		log.Println("========== !!! 背景程序摧毀完畢 !!! ==========")
	}

	schedule.Running = false
}
