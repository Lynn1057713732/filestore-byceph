package handler

import (
	dao "filestore-byceph/db"
	"filestore-byceph/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	redisPool "filestore-byceph/cache"
)


//分块的初始化信息
type MultipartUploadInfo struct {
	FileHash string
	FileSize int
	UploadID string  //分块的唯一id，每一次上传都会有不同的id
	ChunkSize int  //分块大小
	ChunkCount int  //分块总数
}


//初始化分块上传
func InitialMultipartUploadHandler(c *gin.Context) {

	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize, err := strconv.Atoi(c.Request.FormValue("filesize"))
	if err != nil{
		c.Data(http.StatusOK, "application/json",
			utils.NewRespMsg(
				-1, "Invalid Filesize", nil).JSONBytes())
		return
	}

	//获取一个redis链接
	rConn := redisPool.RedisPool().Get()
	defer rConn.Close()

	//生成分块上传的初始化信息
	upInfo := MultipartUploadInfo{
		FileHash: filehash,
		FileSize: filesize,
		UploadID: username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize: 5 * 1024 * 1024,  //5Mb
		ChunkCount: int(math.Ceil(float64(filesize)) / (5 * 1024 * 1024)),
	}

	//将初始化信息写入到redis中
	rConn.Do("HSET", "MP_" + upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	rConn.Do("HSET", "MP_" + upInfo.UploadID, "filehash", upInfo.FileHash)
	rConn.Do("HSET", "MP_" + upInfo.UploadID, "filesize", upInfo.FileSize)

	//将响应初始化信息返回到客户端
	c.Data(http.StatusOK, "application/json",
		utils.NewRespMsg(0, "OK", upInfo).JSONBytes())


}

// UploadPartHandler : 上传文件分块
func UploadPartHandler(c * gin.Context) {
	// 1. 解析用户请求参数
	//	username := r.Form.Get("username")
	uploadID := c.Request.FormValue("uploadid")
	chunkIndex := c.Request.FormValue("index")

	// 2. 获得redis连接池中的一个连接
	rConn := redisPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 获得文件句柄，用于存储分块内容
	fpath := "/Users/yangfengming/opt/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		c.Data(http.StatusOK, "application/json",
			utils.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
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

	// 4. 更新redis缓存状态
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	// 5. 返回处理结果到客户端
	c.Data(http.StatusOK, "application/json",
		utils.NewRespMsg(0, "OK", nil).JSONBytes())
}

// CompleteUploadHandler : 通知上传合并
func CompleteUploadHandler(c * gin.Context) {
	// 1. 解析请求参数
	upid := c.Request.FormValue("uploadid")
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filesize := c.Request.FormValue("filesize")
	filename := c.Request.FormValue("filename")

	// 2. 获得redis连接池中的一个连接
	rConn := redisPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 通过uploadid查询redis并判断是否所有分块上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+upid))
	if err != nil {
		c.Data(http.StatusOK, "application/json",
			utils.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
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
		c.Data(http.StatusOK, "application/json",
			utils.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}

	// 4. TODO：合并分块

	// 5. 更新唯一文件表及用户文件表
	fsize, _ := strconv.Atoi(filesize)
	dao.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	dao.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	// 6. 响应处理结果
	c.Data(http.StatusOK, "application/json",
		utils.NewRespMsg(0, "OK", nil).JSONBytes())

}

func CancelUploadPartHandler(c * gin.Context) {
	//删除已存在的分块文件
	//删除redis缓存分块
	//更新mysql文件status

}



func MultipartUploadStatusHandler(c * gin.Context) {
	//检查分块上传状态是否有效
	//获取分块初始化信息
	//获取已上传的分块信息

}
