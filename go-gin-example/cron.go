/**
 * 定时任务（这里做的是定时清理数据库中被软删除的数据）
 * 使用方式：go run cron.go
 */
package main

import (
	"log"
	"time"

	"github.com/PeiLeizzz/go-gin-example/models"
	"github.com/robfig/cron"
)

func main() {
	log.Println("Starting...")

	c := cron.New()
	c.AddFunc("* * * * * *", func() {
		log.Println("Run models.CleanAllTags...")
		models.CleanAllTags()
	})
	c.AddFunc("* * * * * *", func() {
		log.Println("Run models.CleanAllArticles...")
		models.CleanAllArticles()
	})

	c.Start()

	// 阻塞主程序，直接 select {} 也是可以的
	t1 := time.NewTimer(time.Second * 10)
	for {
		select {
		case <-t1.C:
			t1.Reset(time.Second * 10)
		}
	}
}
