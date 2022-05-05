package main

import (
	"fmt"
	"log"
	"syscall"

	"github.com/PeiLeizzz/go-gin-example/models"
	"github.com/PeiLeizzz/go-gin-example/pkg/gredis"
	"github.com/PeiLeizzz/go-gin-example/pkg/logging"
	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
	"github.com/PeiLeizzz/go-gin-example/routers"
	"github.com/fvbock/endless"
)

func main() {
	// 初始化
	setting.Setup()
	models.Setup()
	logging.Setup()
	gredis.Setup()

	defer func() {
		models.CloseDB()
	}()

	// endless 热更新方法 ↓
	endless.DefaultReadTimeOut = setting.ServerSetting.ReadTimeout
	endless.DefaultWriteTimeOut = setting.ServerSetting.WriteTimeout
	endless.DefaultMaxHeaderBytes = 1 << 20
	endPort := fmt.Sprintf(":%d", setting.ServerSetting.HttpPort)

	server := endless.NewServer(endPort, routers.InitRouter())
	server.BeforeBegin = func(add string) {
		log.Printf("Actual pid is %d", syscall.Getpid())
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Server err: %v", err)
	}

	// http.Server 的 shutdown 方法 ↓
	// router := routers.InitRouter()

	// s := &http.Server{
	// 	Addr:           fmt.Sprintf(":%d", setting.HTTPPort),
	// 	Handler:        router,
	// 	ReadTimeout:    setting.ReadTimeout,
	// 	WriteTimeout:   setting.WriteTimeout,
	// 	MaxHeaderBytes: 1 << 20,
	// }

	// s.Shutdown 一旦执行，s.ListenAndServe 就会退出，所以要单独放在另一个 goroutine 中
	// go func() {
	// 	if err := s.ListenAndServe(); err != nil {
	// 		log.Printf("Listen: %s\n", err)
	// 	}
	// }()

	// quit := make(chan os.Signal)
	// signal.Notify(quit, os.Interrupt)
	// <-quit

	// log.Println("Shutdown Server ...")

	// 传入父 Context，返回子 Context，超过设定时间默认视为子 Context 相关操作完成（自动调用 cancel()）
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel() // 子 Context 操作完成后应立即调用取消函数
	// if err := s.Shutdown(ctx); err != nil {
	// 	log.Fatal("Server shutdown:", err)
	// }

	// log.Println("Server exiting")
}
