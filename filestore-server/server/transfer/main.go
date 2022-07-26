package main

import (
	"bufio"
	"encoding/json"
	"filestore-server/config"
	"filestore-server/db"
	"filestore-server/mq"
	"filestore-server/store/oss"
	"log"
	"os"
)

func ProcessTransfer(msg []byte) bool {
	// 解析 msg
	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 根据临时存储文件路径，创建文件句柄
	file, err := os.Open(pubData.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 通过文件句柄将文件内容读出来并且上传到 oss
	err = oss.Bucket().PutObject(pubData.DestLocation, bufio.NewReader(file))
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 更新文件的存储路径到文件表
	ok, err := db.UpdateFileLocation(pubData.FileHash, pubData.DestLocation)
	if !ok || err != nil {
		log.Println("Failed to update file location")
		return false
	}
	// 删除临时文件
	_ = os.Remove(pubData.CurLocation) // ignore error
	return true
}

func main() {
	if !config.AsyncTransferEnable {
		log.Println("未开启异步配置")
		return
	}

	log.Println("开始监听转移任务队列...")
	mq.StartConsume(config.TransOSSQueueName, "transfer_oss", ProcessTransfer)
}
