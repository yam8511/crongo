package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/tucnak/telebot"
	"github.com/yam8511/crongo"
)

func main() {
	var err interface{}
	/// 設定參數
	envFile := flag.String("e", "", "指定 env 檔案名稱")
	flag.Parse()

	/// 讀取 ENV 設定檔
	if *envFile == "" {
		err = godotenv.Load(".env")
	} else {
		err = godotenv.Load(*envFile)
	}
	CheckErrFatal(err, "讀取 env 錯誤")

	// Telegram - 讀取設定檔，建立機器人
	BotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	cahtid, err := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
	CheckErrFatal(err, "<BOT_TOKEN> 格式錯誤")
	ChatID := int64(cahtid)

	/// Telegram - 伺服器關掉時通知
	AdminChat := telebot.Chat{ID: ChatID}
	Bot, err := telebot.NewBot(BotToken)
	CheckErrFatal(err, "建立機器人錯誤")

	// 程式結束時，通知背景已關閉
	defer func(Bot *telebot.Bot) {
		message := fmt.Sprintf("【 %s : %s 】 排程已關閉！", os.Getenv("PROJECT_ENV"), os.Getenv("MACHINE_IP"))
		if err = recover(); err != nil {
			log.Printf("[Error] %v", err)
			message += fmt.Sprintf(" (%v)", err)
		}
		Bot.SendMessage(AdminChat, message, nil)
	}(Bot)

	/// 宣告系統信號
	sigs := make(chan os.Signal, 1)
	exit := make(chan interface{})
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	/// 設置 Ctrl + C 機制
	go func() {
		log.Println("[Info] 結束背景請按 <Ctrl + C>")
		// 等待 Ctrl + C 的信號
		// 離開程式
		exit <- <-sigs
	}()

	// 開始排程
	schdule := crongo.NewSchedule()
	one := schdule.NewShell("Zuolar", "* * * * * *", "touch", []string{"1"}, false)
	schdule.AddMission(one.Cron, one)
	schdule.Run()

	/// 接收信號，結束程式
	log.Printf("[Info] Receive Signal: %v", <-exit)
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
