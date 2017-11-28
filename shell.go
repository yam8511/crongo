package crongo

import (
	"log"
	"os"
	"os/exec"
)

// Shell : Shell指令
type Shell struct {
	// 任務名稱
	Name string `json:"name"`
	// 執行週期
	Cron string `json:"cron"`
	// 指令
	Command string `json:"command"`
	// 指令參數
	Args []string `json:"args"`
	// 是否能重複行
	Overlapping bool `json:"overlapping"`
	// 已執行的PIDs
	Pids []int `json:"pids"`
}

// Run : 執行任務
func (shell *Shell) Run() {
	// 若 Overlapping is False, 先檢查有沒有已經執行的程序
	// 若已經有執行的程序，則不執行
	if !shell.Overlapping && len(shell.Pids) > 0 {
		return
	}

	// 模仿 Terminal 輸入指令
	cmd := exec.Command(shell.Command, shell.Args...)

	// 載入系統的環境變數
	cmd.Env = os.Environ()

	// 模仿 Terminal 按下 Enter 鍵
	err := cmd.Start()
	// 如果有錯誤，則結束程式並且印出錯誤訊息
	if err != nil {
		log.Fatalf("[Error] Command: <%s> start with error %v\n", shell.Name, err)
	}

	// 記下程序的PID
	shell.Pids = append(shell.Pids, cmd.Process.Pid)

	// Debug用
	// log.Println("Name: ", shell.Name, " \nPID:", shell.Pids)

	// 等待 command 執行結束
	err = cmd.Wait()
	if err != nil {
		log.Fatalf("[Error] Command: <%s> wait with error %v\n", shell.Name, err)
	}

	// 清除該程序的PID
	index := indexOf(shell.Pids, cmd.Process.Pid)
	if index != -1 {
		shell.Pids = append(shell.Pids[:index], shell.Pids[index+1:]...)
	}

	// Debug用
	// log.Println("Name: ", shell.Name, " \nPID:", shell.Pids)
	log.Printf("[Info] Command: <%s> finished with error: %v\n", shell.Name, err)
}
