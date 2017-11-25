package main

import (
	"log"
	"os/exec"
)

func main() {
	// 程式離開前，最後一項任務
	defer func() {
		log.Println("Finish!")
	}()

	cmd := exec.Command("sleep", "15")
	// 如果用Run，执行到该步则会阻塞等待5秒
	// err := cmd.Run()
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("PID:", cmd.Process.Pid)
	log.Printf("Waiting for command to finish...")
	// Start，上面的内容会先输出，然后这里会阻塞等待5秒
	err = cmd.Wait()
	log.Printf("Command finished with error: %v", err)

	// schedule := new(Schedule)
	// schedule.Run()
}
