# CronGo

```shell
go get -u github.com/kardianos/govendor
go get -u github.com/yam8511/crongo
```

---

## 執行範例
```shell
$ cd $GOPATH/src/
$ cp -R $GOPATH/src/github.com/yam8511/crongo/example/telegram_signal cronjob
$ cd cronjob
$ govendor sync
$ go build .
$ ./cronjob

2017/11/30 04:12:52 ====== !!! 開始啟動背景程序 !!! ======
2017/11/30 04:12:52 [Info] 背景已啟動，結束背景請按 <Ctrl + C>
2017/11/30 04:12:53 [Info] Name: Zuolar-Touch , PID: [17211]
2017/11/30 04:12:53 [OK] Command:〈 Zuolar-Touch 〉, PID:〈 17211 〉, Finish with error: <nil>
2017/11/30 04:12:54 [Info] Name: Zuolar-Touch , PID: [17212]
2017/11/30 04:12:54 [Info] Name: Zuolar-Remove , PID: [17213]
2017/11/30 04:12:54 [OK] Command:〈 Zuolar-Touch 〉, PID:〈 17212 〉, Finish with error: <nil>
2017/11/30 04:12:54 [OK] Command:〈 Zuolar-Remove 〉, PID:〈 17213 〉, Finish with error: <nil>
2017/11/30 04:12:55 [Info] Name: Zuolar-Touch , PID: [17214]
2017/11/30 04:12:55 [OK] Command:〈 Zuolar-Touch 〉, PID:〈 17214 〉, Finish with error: <nil>
< Ctrl + C >
2017/11/30 04:12:55 [Warning] Receive Signal: interrupt
2017/11/30 04:12:55 ======= 等待背景以下程序結束... =======
2017/11/30 04:12:55 =======================================
2017/11/30 04:12:55 ======== !!! 背景程序已暫停 !!! =======
2017/11/30 04:12:55 [Info] 程式結束
```
