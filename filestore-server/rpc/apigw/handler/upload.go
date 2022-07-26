package handler

import (
	"encoding/json"
	"filestore-server/common"
	globalConfig "filestore-server/config"
	"filestore-server/db"
	"filestore-server/meta"
	"filestore-server/mq"
	"filestore-server/store/ceph"
	"filestore-server/store/oss"
	"filestore-server/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/amz.v1/s3"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

func PostUploadHandler(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")

	// 接收文件流、存储到本地目录
	head, err := c.FormFile("file")
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to get header, err: %s\n", err.Error()))
		return
	}
	file, err := head.Open()
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to get data, err: %s\n", err.Error()))
		return
	}
	defer file.Close()

	// 创建文件元信息
	fileMeta := &meta.FileMeta{
		FileName: head.Filename,
		Location: "./tmp/" + head.Filename,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	// 创建本地文件
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to create file, err: %s\n", err.Error()))
		return
	}
	defer newFile.Close()

	// 写入到本地文件中
	fileMeta.FileSize, err = io.Copy(newFile, file)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to save data into new file, err: %s\n", err.Error()))
		return
	}

	// 计算 sha1 值
	newFile.Seek(0, 0) // 移到文件开头 ignore err
	fileMeta.FileSha1 = util.FileSha1(newFile)

	if globalConfig.CurrentStoreType == common.StoreCeph {
		// 将文件写入 ceph
		newFile.Seek(0, 0) // ignore err
		data, err := ioutil.ReadAll(newFile)
		if err != nil {
			internelServerError(c, fmt.Sprintf("Failed to read the new file, err: %s\n", err.Error()))
			return
		}
		bucket := ceph.GetCephBucket("userfile")
		cephPath := "/ceph/" + fileMeta.FileSha1
		err = bucket.Put(cephPath, data, "octet-stream", s3.PublicRead)
		if err != nil {
			internelServerError(c, fmt.Sprintf("Failed to save into ceph, err: %s\n", err.Error()))
			return
		}
		fileMeta.Location = cephPath
	} else if globalConfig.CurrentStoreType == common.StoreOSS {
		// 将文件写入 oss
		newFile.Seek(0, 0) // ignore err
		ossPath := "oss/" + fileMeta.FileSha1

		if !globalConfig.AsyncTransferEnable {
			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				internelServerError(c, fmt.Sprintf("Failed to save into oss, err: %s\n", err.Error()))
				return
			}
			fileMeta.Location = ossPath
		} else {
			data := mq.TransferData{
				FileHash:      fileMeta.FileSha1,
				CurLocation:   fileMeta.Location,
				DestLocation:  ossPath,
				DestStoreType: common.StoreOSS,
			}
			pubData, _ := json.Marshal(data)
			pubSuc := mq.Publish(
				globalConfig.TransExchangeName,
				globalConfig.TransOSSRoutingKey,
				pubData,
			)
			if !pubSuc {
				internelServerError(c, fmt.Sprintf("Failed to save into oss(rabbitmq), err: %s\n", err.Error()))
				return
			}
		}
	}

	// 向数据库添加 fileMetas 元信息
	_, err = meta.InsertFileMetaDB(fileMeta) // ignore bool
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to save meta data, err: %s\n", err.Error()))
		return
	}

	// 更新用户文件表记录
	username := c.PostForm("username")
	if username == "" {
		username = c.Query("username")
	}
	_, err = db.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize) // ignore bool
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to save user_info data, err: %s\n", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "success",
	})
}

func internelServerError(c *gin.Context, msg string) {
	c.String(http.StatusInternalServerError, "%s", msg)
	return
}
