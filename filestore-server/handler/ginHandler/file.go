package ginHandler

import (
	"encoding/json"
	"filestore-server/common"
	"filestore-server/config"
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
	"strconv"
	"time"
)

func GetUploadHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/index.html")
}

func PostUploadHandler(c *gin.Context) {
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

	if config.CurrentStoreType == common.StoreCeph {
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
	} else if config.CurrentStoreType == common.StoreOSS {
		// 将文件写入 oss
		newFile.Seek(0, 0) // ignore err
		ossPath := "oss/" + fileMeta.FileSha1

		if !config.AsyncTransferEnable {
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
				config.TransExchangeName,
				config.TransOSSRoutingKey,
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

	c.Redirect(http.StatusFound, "/file/upload/success")
}

func PostTryFastUploadHandler(c *gin.Context) {
	username := c.PostForm("username")
	filehash := c.PostForm("filehash")
	filename := c.PostForm("filename")
	filesize, err := strconv.ParseInt(c.PostForm("filesize"), 10, 64)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to parse filesize, err: %s\n", err.Error()))
		return
	}

	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to get fileMeta, err: %s\n", err.Error()))
		return
	} else if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		data, err := json.Marshal(resp)
		if err != nil {
			internelServerError(c, fmt.Sprintf("Failed to marshal fileMeta, err: %s\n", err.Error()))
			return
		}

		c.String(http.StatusOK, "%s", string(data))
	}

	// 写入用户文件表
	// TODO: 如果该用户已经上传过，则改为修改用户文件表（目前是允许重复写入同一个文件）
	_, err = db.OnUserFileUploadFinished(username, filehash, filename, filesize) // ignore bool
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to save user_info data, err: %s\n", err.Error()))
		return
	}

	c.Redirect(http.StatusFound, "/file/upload/success")
}

func GetUploadSuccessHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain", []byte("Upload success!"))
}

func GetFileMetaHandler(c *gin.Context) {
	filehash := c.Query("filehash")
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to get fileMeta, err: %s\n", err.Error()))
		return
	}

	data, err := json.Marshal(fMeta)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to marshal fileMeta, err: %s\n", err.Error()))
		return
	}

	c.String(http.StatusOK, "%s", string(data))
}

func GetFileQueryHandler(c *gin.Context) {
	username := c.PostForm("username")
	limit, err := strconv.Atoi(c.PostForm("limit"))
	if err != nil {
		c.String(http.StatusBadRequest, "%s", "illegal limit parameter")
		return
	}

	userFiles, err := db.QueryUserFileMetas(username, limit)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to query user file metas, err: %s\n", err.Error()))
		return
	}

	data, err := json.Marshal(userFiles)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to marshal user file meta data, err: %s\n", err.Error()))
		return
	}

	c.String(http.StatusOK, "%s", string(data))
}

func GetDownloadHandler(c *gin.Context) {
	filehash := c.Query("filehash")

	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to get fileMeta, err: %s\n", err.Error()))
		return
	} else if fMeta == nil {
		internelServerError(c, fmt.Sprintf("No such file, sha1: %s\n", filehash))
		return
	}

	// TODO: 这里直接读取后以字节流形式发给前端；更好的做法是通过静态服务器（nginx 等）以静态资源的形式给用户
	f, err := os.Open(fMeta.Location)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to open file, err: %s\n", err.Error()))
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to read file, err: %s\n", err.Error()))
		return
	}

	// 改写 http 头，使浏览器识别后直接下载
	c.Header("Content-Type", "application/octect-stream")
	c.Header("content-disposition", `attachment; filename="`+fMeta.FileName+`"`)
	c.String(http.StatusOK, "%s", data)
}

func GetDownloadURLHandler(c *gin.Context) {
	filehash := c.Query("filehash")

	// 从文件表中查找记录
	fileMeta, err := db.GetFileMeta(filehash)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to get file meta, err: %s\n", err))
		return
	}

	// TODO: 判断文件存在 OSS 还是 ceph

	signedURL := oss.DownloadURL(fileMeta.FileAddr.String)
	c.String(http.StatusOK, "%s", signedURL)
}

func PostFileMetaUpdateHandler(c *gin.Context) {
	opType := c.PostForm("op")
	filehash := c.PostForm("filehash")
	newFileName := c.PostForm("filename")
	if opType != "0" {
		c.String(http.StatusForbidden, "%s", "This operator type is illegal")
		return
	}

	// 获取原来的文件元信息
	curFileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to get fileMeta, err: %s\n", err.Error()))
		return
	} else if curFileMeta == nil {
		c.String(http.StatusBadRequest, "%s", "No such file, sha1: "+filehash)
		return
	}

	// 更新文件名称，并且在元信息集合中更新
	preFileName := curFileMeta.FileName
	curFileMeta.FileName = newFileName
	ok, err := meta.UpdateFileMetaDB(curFileMeta)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to update file meta, err: %s\n", err.Error()))
		return
	} else if !ok {
		c.String(http.StatusBadRequest, "%s", "No such file: "+preFileName)
		return
	}

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to marshal fileMeta, err: %s\n", err.Error()))
		return
	}

	c.String(http.StatusOK, "%s", string(data))
}

func GetFileDeleteHandler(c *gin.Context) {
	filehash := c.Query("filehash")

	// 删除文件本体
	// TODO：改成向 DB 请求
	fMeta := meta.GetFileMeta(filehash)
	if fMeta == nil {
		internelServerError(c, fmt.Sprintf("No such file, sha1: %s\n", filehash))
		return
	}
	_ = os.Remove(fMeta.Location) // ignore errors

	// 删除文件元信息（索引）
	meta.RemoveFileMeta(filehash)

	c.String(http.StatusOK, "%s", "delete success")
}
