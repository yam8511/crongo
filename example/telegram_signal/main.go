package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/yam8511/crongo"
	"gopkg.in/telegram-bot-api.v4"
)

func main() {
	var err interface{}
	/// 設定參數
	envFile := flag.String("env", ".env", "指定 env 檔案名稱")
	jsonFile := flag.String("json", "cron.json", "指定「排程背景工作」的 json檔案")
	flag.Parse()

	/// 讀取 ENV 設定檔
	err = godotenv.Load(*envFile)
	CheckErrFatal(err, "讀取 env 錯誤")

	// Telegram - 讀取設定檔
	BotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	cahtid, err := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
	CheckErrFatal(err, "<TELEGRAM_BOT_TOKEN> 格式錯誤")
	AdminChatID := int64(cahtid)

	// Telegram - 建立機器人
	Bot, err := tgbotapi.NewBotAPI(BotToken)
	CheckErrFatal(err, "建立機器人錯誤")
	MessageStack := map[string]time.Time{}
	BotSendMessage := func(message string) {
		MessageStack[message] = time.Now()
	}

	// Telegram - 定時發送訊息
	const noticeSecond = time.Second * 60 // 幾秒鐘通知一次
	ticker := time.NewTicker(noticeSecond)
	go func() {
		for range ticker.C {
			for message, t := range MessageStack {
				message = fmt.Sprintf("【 %s : %s 】\n時間：%s\n%s", os.Getenv("PROJECT_ENV"), os.Getenv("MACHINE_IP"), t.Format("2006-01-02 03:04:05"), message)
				msg := tgbotapi.NewMessage(AdminChatID, message)
				Bot.Send(msg)
			}
			MessageStack = map[string]time.Time{}
		}
	}()

	// 程式結束時，通知背景已關閉
	defer func(BotSendMessage func(string)) {
		message := fmt.Sprintf("【 %s : %s 】 排程已關閉！", os.Getenv("PROJECT_ENV"), os.Getenv("MACHINE_IP"))
		if err = recover(); err != nil {
			log.Printf("[Error] %v", err)
			message += fmt.Sprintf(" (%v)", err)
		}
		msg := tgbotapi.NewMessage(AdminChatID, message)
		Bot.Send(msg)
	}(BotSendMessage)

	/// 宣告系統信號
	sigs := make(chan os.Signal, 1)
	exit := make(chan interface{})
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// 設置 Ctrl + C 機制
	go func() {
		log.Println("[Info] 背景已啟動，結束背景請按 <Ctrl + C>")
		// 等待 Ctrl + C 的信號
		// 離開程式
		exit <- <-sigs
	}()

	// 通知開啟背景
	go func(BotSendMessage func(string)) {
		message := fmt.Sprintf("【 %s : %s 】 排程已開啟！", os.Getenv("PROJECT_ENV"), os.Getenv("MACHINE_IP"))
		msg := tgbotapi.NewMessage(AdminChatID, message)
		Bot.Send(msg)
	}(BotSendMessage)

	// ----------- 解析「排程工作」內容 -------------
	jobsJSON, err := ioutil.ReadFile(*jsonFile)
	CheckErrFatal(err, "讀取〈"+(*jsonFile)+"〉錯誤")
	missions := []*crongo.Shell{}
	json.Unmarshal(jobsJSON, &missions)

	// ----------- 開始排程 -------------
	schdule := crongo.NewSchedule()
	for _, mission := range missions {
		mission.ErrorHandler = func(err error) {
			log.Printf("【 %s : %s 】 Command:〈 %s 〉throw error %v！\n", os.Getenv("PROJECT_ENV"), os.Getenv("MACHINE_IP"), mission.Name, err)
			message := fmt.Sprintf("【 %s : %s 】 Command:〈 %s 〉throw error %v！", os.Getenv("PROJECT_ENV"), os.Getenv("MACHINE_IP"), mission.Name, err)
			BotSendMessage(message)
		}

		schdule.AddMission(mission.Cron, mission)
	}
	schdule.Run()

	// 接收信號，先停止背景程序
	log.Printf("[Warning] Receive Signal: %v", <-exit)
	schdule.Suspend()
	// 結束程式
	log.Println("[Info] 程式結束")
}

// CheckErrFatal : 確認錯誤，如果有錯誤則結束程式
func CheckErrFatal(err interface{}, msg ...interface{}) {
	if err != nil {
		if len(msg) == 0 {
			log.Fatalf("[Error] %v", err)
		}
		log.Fatalln("[Error]", msg, err)
	}
}
