package crongo

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// Shell : Shell指令
type Shell struct {
	// 任務名稱
	Name string `toml:"name,omitempty" json:"name"`
	// 執行週期
	Cron string `toml:"cron,omitempty" json:"cron"`
	// 指令
	Command string `toml:"command,omitempty" json:"command"`
	// 指令參數
	Args []string `toml:"args,omitempty" json:"args"`
	// 環境參數
	Env []string `toml:"env,omitempty" json:"env"`
	// 是否能重複行
	Overlapping bool `toml:"overlapping,omitempty" json:"overlapping"`
	// 是否常駐
	Permanent bool `toml:"permanent,omitempty" json:"permanent"`
	// 已執行的PIDs
	Pids []int `toml:"pids,omitempty" json:"pids"`
	// 是否啟動
	IsEnable bool `toml:"enable,omitempty" json:"enable"`
	// 錯誤處理方式
	ErrorHandler func(*exec.Cmd, error)
	// 前置作業事件
	PrepareHandler func(*exec.Cmd) error
	// 作業完成事件
	FinishHandler func(*exec.Cmd)
	// 讀寫鎖
	mutex *sync.RWMutex
}

// Run : 執行任務
func (shell *Shell) Run() {
	defer func() {
		if err := recover(); err != nil {
			writeLog(fmt.Sprintf("[ERROR] Task〈%s〉unexpected error (%v)", shell.Name, err))
			return
		}
	}()

	// 模仿 Terminal 輸入指令
	cmd := exec.Command(shell.Command, shell.Args...)

	// 載入系統的環境變數
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, shell.Env...)
	if shell.PrepareHandler != nil {
		preErr := shell.PrepareHandler(cmd)
		if preErr != nil {
			shell.ErrorHandler(cmd, preErr)
		}
	}

	// 若任務沒有啟動，則不執行
	if !shell.IsEnable {
		return
	}
	// 若 Overlapping is False, 先檢查有沒有已經執行的程序
	// 若已經有執行的程序，則不執行
	if !shell.Overlapping && len(shell.Pids) > 0 {
		return
	}

	// 模仿 Terminal 按下 Enter 鍵
	err := cmd.Start()
	// 如果有錯誤，則結束程式並且執行錯誤處理
	if err != nil && shell.ErrorHandler != nil {
		shell.ErrorHandler(cmd, err)
		return
	}

	// 記下程序的PID
	shell.mutex.Lock()
	shell.Pids = append(shell.Pids, cmd.Process.Pid)
	shell.mutex.Unlock()

	// Debug用
	writeLog(fmt.Sprintf("[INFO] Task〈%s〉Start with PID #%d", shell.Name, cmd.Process.Pid))

	// 等待 command 執行結束
	err = cmd.Wait()
	if err != nil && shell.ErrorHandler != nil {
		shell.ErrorHandler(cmd, err)
	}

	// 執行結束的動作
	if shell.FinishHandler != nil {
		shell.FinishHandler(cmd)
	}

	// 清除該程序的PID
	shell.mutex.Lock()
	index := indexOf(shell.Pids, cmd.Process.Pid)
	if index != -1 {
		shell.Pids = append(shell.Pids[:index], shell.Pids[index+1:]...)
	}
	shell.mutex.Unlock()

	// Debug用
	var exitCode interface{}
	if err != nil {
		exitCode = err
	} else {
		exitCode = 0
	}
	writeLog(fmt.Sprintf("[INFO] Task〈 %s 〉#%d exit (%v)", shell.Name, cmd.Process.Pid, exitCode))
}

// Enable : 開啟任務
func (shell *Shell) Enable() {
	shell.IsEnable = true
}

// Disable : 關閉任務
func (shell *Shell) Disable() {
	shell.IsEnable = false
}

// GetPids : 取目前執行的PID
func (shell *Shell) GetPids() []int {
	return shell.Pids
}

// GetName : 取目前任務的名稱
func (shell *Shell) GetName() string {
	return shell.Name
}

// GetCron : 取目前任務的排程時間
func (shell *Shell) GetCron() string {
	return shell.Cron
}

// IsPermanent : 是否為常駐程序
func (shell *Shell) IsPermanent() bool {
	return shell.Permanent
}
