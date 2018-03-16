package crongo

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	cron "gopkg.in/robfig/cron.v2"
)

// Schedule : 背景排程
type Schedule struct {
	// Missions : 需要執行的背景任務
	Missions []Mission
	Running  bool
	Cron     *cron.Cron
	mutex    *sync.RWMutex
}

// Mission : 任務介面
type Mission interface {
	GetCron() string
	Run()
	Enable()
	Disable()
	GetName() string
	GetPids() []int
	IsPermanent() bool
}

// NewSchedule : 建立一個新排程
func NewSchedule() *Schedule {
	newSchedule := &Schedule{
		Missions: []Mission{},
		Running:  false,
		Cron:     cron.New(),
		mutex:    new(sync.RWMutex),
	}
	return newSchedule
}

// NewShell : 建立一個新任務
func (schedule *Schedule) NewShell(
	name, cron, command string,
	args, env []string,
	overlapping, permanet, enable bool,
	errorHandler func(*exec.Cmd, error) error,
	prepareHandler func(*exec.Cmd) error,
	finishHandler func(*exec.Cmd) error,
) *Shell {
	return &Shell{
		Name:           name,
		Cron:           cron,
		Command:        command,
		Args:           args,
		Env:            env,
		Overlapping:    overlapping,
		Permanent:      permanet,
		Pids:           []int{},
		IsEnable:       enable,
		ErrorHandler:   errorHandler,
		PrepareHandler: prepareHandler,
		FinishHandler:  finishHandler,
		mutex:          new(sync.RWMutex),
	}
}

// AddMission : 新增任務到背景排程
func (schedule *Schedule) AddMission(mission Mission) error {
	schedule.mutex.Lock()
	schedule.Missions = append(schedule.Missions, mission)
	schedule.mutex.Unlock()
	_, err := schedule.Cron.AddJob(mission.GetCron(), mission)
	return err
}

// Run : 開始執行背景排程
func (schedule *Schedule) Run() {
	// 如果背景正在跑，則跳過
	if schedule.Running {
		writeLog("======= 目前背景排程已經啟動... =======")
		return
	}

	writeLog("====== !!! 開始啟動背景程序 !!! ======")
	schedule.Running = true
	schedule.Cron.Start()
}

// Suspend : 停止背景排程
func (schedule *Schedule) Suspend() {
	if !schedule.Running {
		writeLog("======= 目前背景排程已經停止... =======")
		return
	}

	schedule.Cron.Stop()

	killPids := []string{}
	writeLog("======= 等待背景以下程序結束... =======")
	for _, mission := range schedule.Missions {
		pids := mission.GetPids()
		if mission.IsPermanent() {
			for _, pid := range pids {
				killPids = append(killPids, fmt.Sprint(pid))
			}
		}
		if len(pids) > 0 {
			writeLog(fmt.Sprintf("> Command:〈 %s 〉, PID:〈 %d 〉", mission.GetName(), pids))
		}
	}
	writeLog("=======================================")

	if len(killPids) > 0 {
		writeLog("========== !!! 常駐程序摧毀開始 !!! ==========")
		killer := exec.Command("kill", killPids...)
		stdin := strings.Join(killer.Args, " ")
		writeLog(fmt.Sprintf("[STDIN] %s", stdin))

		killer.Stdout = bytes.NewBuffer(nil)
		killer.Stderr = bytes.NewBuffer(nil)
		err := killer.Run()
		time.Sleep(time.Second)
		if err != nil {
			stdout := string(killer.Stdout.(*bytes.Buffer).Bytes())
			stderr := string(killer.Stderr.(*bytes.Buffer).Bytes())
			writeLog("========== !!! 常駐程序摧毀發生錯誤 !!! ==========")
			writeLog(fmt.Sprintf("[STDOUT] %v", stdout))
			writeLog(fmt.Sprintf("[STDERR] %v", stderr))
		} else {
			writeLog("========== !!! 常駐程序摧毀完畢 !!! ==========")
		}
	}

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
	writeLog("======== !!! 背景程序已暫停 !!! =======")
}

// Destroy : 強制停止背景排程
func (schedule *Schedule) Destroy() {
	if !schedule.Running {
		writeLog("======== 目前背景排程尚未啟動... ========")
		return
	}
	schedule.Cron.Stop()

	writeLog("======== !!! 即將摧毀以下背景程序 !!! ========")
	for _, mission := range schedule.Missions {
		if pids := mission.GetPids(); len(pids) > 0 {
			writeLog(fmt.Sprintf("> Task〈 %s 〉, PID:〈 %v 〉", mission.GetName(), pids))
		}
	}
	writeLog("==============================================")

	timer := time.NewTicker(time.Second)
	tick := 5
	go func() {
		for range timer.C {
			writeLog(fmt.Sprintf("[WARNING] 倒數%d秒.....", tick))
			tick--
		}
	}()
	for tick > 0 {
	}
	timer.Stop()

	writeLog("========== !!! 開始摧毀背景程序 !!! ==========")
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
		writeLog("========== !!! 背景摧毀發生錯誤 !!! ==========")
		writeLog(fmt.Sprintf("[ERROR] %v", err))
	} else {
		writeLog("========== !!! 背景程序摧毀完畢 !!! ==========")
	}

	schedule.Running = false
}
