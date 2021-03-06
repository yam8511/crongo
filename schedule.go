package crongo

import (
	"fmt"
	"os/exec"
	"sync"
	"time"

	cron "gopkg.in/robfig/cron.v2"
)

// Schedule : 背景排程
type Schedule struct {
	// CronJobs : 需要執行的背景任務
	CronJobs map[string]CronJob
	// Missions []Mission
	Running bool
	Cron    *cron.Cron
	mutex   *sync.RWMutex
}

// CronJob 背景工作
type CronJob struct {
	Mission Mission
	EntryID cron.EntryID
}

// Mission : 任務介面
type Mission interface {
	GetCron() string
	Run()
	Stop()
	Enable()
	Disable()
	GetName() string
	GetPids() []int
	IsPermanent() bool
	IsRunning() bool
}

// NewSchedule : 建立一個新排程
func NewSchedule() *Schedule {
	newSchedule := &Schedule{
		CronJobs: map[string]CronJob{},
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
	defer schedule.mutex.Unlock()

	if _, ok := schedule.CronJobs[mission.GetName()]; ok {
		return fmt.Errorf("Job'name has been declared -> %s", mission.GetName())
	}
	entry, err := schedule.Cron.AddJob(mission.GetCron(), mission)
	schedule.CronJobs[mission.GetName()] = CronJob{
		Mission: mission,
		EntryID: entry,
	}
	return err
}

// RemoveMission : 從背景排程刪除任務
func (schedule *Schedule) RemoveMission(name string) (err error) {
	schedule.mutex.Lock()
	defer schedule.mutex.Unlock()
	defer func() {
		if catchErr := recover(); catchErr != nil {
			err = fmt.Errorf("%v", catchErr)
			return
		}
	}()
	if job, ok := schedule.CronJobs[name]; ok {
		schedule.Cron.Remove(job.EntryID)
		job.Mission.Stop()
	}
	return
}

// Start : 開始執行背景排程
func (schedule *Schedule) Start() {
	// 如果背景正在跑，則跳過
	if schedule.Running {
		writeLog("======= 目前背景排程已經啟動... =======")
		return
	}

	writeLog("====== !!! 開始啟動背景程序 !!! ======")
	schedule.Running = true
	schedule.Cron.Start()
}

// Stop : 停止背景排程
func (schedule *Schedule) Stop() {
	if !schedule.Running {
		writeLog("======= 目前背景排程已經停止... =======")
		return
	}

	schedule.Cron.Stop()

	killPids := []string{}
	writeLog("======= 等待背景以下程序結束... =======")
	for _, Job := range schedule.CronJobs {
		Job.Mission.Stop()
		pids := Job.Mission.GetPids()
		if Job.Mission.IsPermanent() {
			for _, pid := range pids {
				killPids = append(killPids, fmt.Sprint(pid))
			}
		}
		if len(pids) > 0 {
			writeLog(fmt.Sprintf("> Command:〈 %s 〉, PID:〈 %d 〉", Job.Mission.GetName(), pids))
		}
	}
	writeLog("=======================================")

	waittingProcessFinish := make(chan int)
	go func() {
		for {
			hasPID := false
			for _, Job := range schedule.CronJobs {
				if len(Job.Mission.GetPids()) > 0 {
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
	for _, Job := range schedule.CronJobs {
		Job.Mission.Stop()
		if pids := Job.Mission.GetPids(); len(pids) > 0 {
			writeLog(fmt.Sprintf("> Task〈 %s 〉, PID:〈 %v 〉", Job.Mission.GetName(), pids))
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

	pids := []int{}
	for _, Job := range schedule.CronJobs {
		for _, pid := range Job.Mission.GetPids() {
			pids = append(pids, pid)
		}
	}

	if len(pids) == 0 {
		writeLog("======== !!! 背景程序已結束 !!! =======")
		return
	}

	writeLog("========== !!! 開始摧毀背景程序 !!! ==========")
	killPids := []string{"-9"}
	for _, pid := range pids {
		killPids = append(killPids, fmt.Sprint(pid))
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
