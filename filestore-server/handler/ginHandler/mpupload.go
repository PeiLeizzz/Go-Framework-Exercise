package ginHandler

import (
	rPool "filestore-server/cache/redis"
	"filestore-server/db"
	"filestore-server/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	DEFAULT_CHUNK_COUNT = 5 * 1024 * 1024
)

type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int64
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// 初始化分块上传(GET)
// TODO 要先判断是否已经传过，如果传过，则秒传
func GetInitialMultipartUploadHandler(c *gin.Context) {
	// 解析用户请求信息
	username := c.Query("username")
	filehash := c.Query("filehash")
	filesize, err := strconv.ParseInt(c.Query("filesize"), 10, 64)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to parse filesize, err: %s\n", err.Error()))
		return
	}

	// 获得 redis 连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 生成分块上传初始化信息
	upInfo := &MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  DEFAULT_CHUNK_COUNT, // 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / DEFAULT_CHUNK_COUNT)),
	}

	// 将初始化信息写入 redis 缓存
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filehash", upInfo.FileHash)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filesize", upInfo.FileSize)

	// 将初始化信息返回客户端
	resp, err := util.NewRespMsg(0, "OK", upInfo).JSONString()
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to marshal upInfo, err: %s\n", err.Error()))
		return
	}
	c.String(http.StatusOK, "%s", resp)
}

// 上传文件分块(GET)
func GetUploadPartHandler(c *gin.Context) {
	// 解析用户请求参数
	uploadID := c.Query("uploadid")
	chunkIndex := c.Query("index")

	// 获得 redis 连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 获得文件句柄，用于存储分块内容
	fpath := "./tmp/data/" + uploadID + "/" + chunkIndex
	_ = os.MkdirAll(path.Dir(fpath), 0744) // ignore
	fd, err := os.Create(fpath)
	if err != nil {
		resp, _ := util.NewRespMsg(-1, "Upload part failed", nil).JSONString() // ignore error
		c.String(http.StatusInternalServerError, "%s", resp)
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := c.Request.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 更新 redis 缓存
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	// 返回处理结果给客户端
	resp, err := util.NewRespMsg(0, "OK", nil).JSONString()
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to marshal upInfo, err: %s\n", err.Error()))
		return
	}
	c.String(http.StatusOK, "%s", resp)
}

// 通知上传合并(GET)
func GetCompleteUploadHandler(c *gin.Context) {
	// 解析请求参数
	username := c.Query("username")
	filehash := c.Query("filehash")
	filename := c.Query("filename")
	uploadID := c.Query("uploadid")
	filesize, err := strconv.ParseInt(c.Query("filesize"), 10, 64)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to parse filesize, err: %s\n", err.Error()))
		return
	}

	// 获取 redis 连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 通过 uploadid 查询 redis 判断是否所有分块上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+uploadID))
	if err != nil {
		resp, _ := util.NewRespMsg(-1, "complete part failed", nil).JSONString() // ignore error
		c.String(http.StatusInternalServerError, "%s", resp)
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		resp, _ := util.NewRespMsg(-2, "invalid request", nil).JSONString() // ignore error
		c.String(http.StatusBadRequest, "%s", resp)
		return
	}

	// TODO：合并分块
	// cat `ls | sort -n` > /tmp/a

	// 更新唯一文件表
	_, _ = db.OnFileUploadFinished(filehash, filename, filesize, "") // ignore error

	// 更新用户文件表
	_, _ = db.OnUserFileUploadFinished(username, filehash, filename, filesize) // ignore error

	// 响应处理结果
	resp, _ := util.NewRespMsg(0, "success", nil).JSONString() // ignore error
	c.String(http.StatusOK, "%s", resp)
	return
}
