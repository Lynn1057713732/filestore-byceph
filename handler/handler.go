package handler

import (
	"encoding/json"
	cmn "filestore-byceph/common"
	"filestore-byceph/mq"
	"filestore-byceph/store/ceph"
	"filestore-byceph/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	cfg "filestore-byceph/config"


	"filestore-byceph/meta"
	dao "filestore-byceph/db"
	"filestore-byceph/store/oss"
)


//UploadHandler:文件上传页面
func UploadHandler(c *gin.Context) {
	c.JSON(http.StatusFound, "/static/view/home.html")


}

func DoUploadHandler (c *gin.Context) {
	errCode := 0
	defer func() {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		if errCode < 0 {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "上传失败",
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "上传成功",
			})
		}
	}()

	//接收文件流及存储到本地目录
	file, head, err  := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Failed to get file data, err:%s\n", err.Error())
		errCode = -2
		return
	}
	defer file.Close()

	//构建文件元信息，并写入存储位置
	fileMeta := meta.FileMeta{
		FileName: head.Filename,
		Location: "/Users/yangfengming/opt/"+head.Filename,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	newFile, err := os.Create(fileMeta.Location)
	if err != nil{
		log.Printf("Failed to create file, err:%s\n", err.Error())
		errCode = -3
	}
	defer newFile.Close()

	fileMeta.FileSize, err = io.Copy(newFile, file)
	if err != nil {
		log.Printf("Failed to save data into file, writtenSize:%d, err:%s\n", fileMeta.FileSize, err.Error())
		errCode = -4
		return
	}

	newFile.Seek(0, 0)
	fileMeta.FileSha1 = utils.FileSha1(newFile)

	// 5. 同步或异步将文件转移到Ceph/OSS
	newFile.Seek(0, 0) // 游标重新回到文件头部
	if cfg.CurrentStoreType == cmn.StoreCeph {
		// 文件写入Ceph存储
		data, _ := ioutil.ReadAll(newFile)
		cephPath := cfg.CephRootDir + fileMeta.FileSha1
		_ = ceph.PutObject("userfile", cephPath, data)
		fileMeta.Location = cephPath
	} else if cfg.CurrentStoreType == cmn.StoreOSS {
		// 文件写入OSS存储
		ossPath := cfg.OSSRootDir + fileMeta.FileSha1
		// 判断写入OSS为同步还是异步
		if !cfg.AsyncTransferEnable {
			// TODO: 设置oss中的文件名，方便指定文件名下载
			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				log.Println(err.Error())
				errCode = -5
				return
			}
			fileMeta.Location = ossPath
		} else {
			// 写入异步转移任务队列
			data := mq.TransferData{
				FileHash:      fileMeta.FileSha1,
				CurLocation:   fileMeta.Location,
				DestLocation:  ossPath,
				DestStoreType: cmn.StoreOSS,
			}
			pubData, _ := json.Marshal(data)
			pubSuc := mq.Publish(
				cfg.TransExchangeName,
				cfg.TransOSSRoutingKey,
				pubData,
			)
			if !pubSuc {
				// TODO: 当前发送转移信息失败，稍后重试
			}
		}
	}

	//meta.UpdateFileMeta(fileMeta)
	_ = meta.CreateFileMetaDB(fileMeta)

	//更新用户文件表
	username := c.Request.FormValue("username")
	suc := dao.OnUserFileUploadFinished(username, fileMeta.FileSha1,
		fileMeta.FileName, fileMeta.FileSize)
	if suc {
		errCode = 0
	} else {
		errCode = -6

	}

}



//UploadSuccessHandler：上传已完成
func UploadSuccessHandler(c *gin.Context) {
	c.JSON(http.StatusOK, "Upload finished!")
}


//GetFileMetaHandler：获取文件元信息
func GetFileMetaHandler(c *gin.Context) {
	log.Println(4444)
	filehash := c.Request.FormValue("filehash")
	log.Println(filehash)
	//fMetaResult := meta.GetFileMeta(filehash)
	fMetaResult, err := meta.GetFileMetaDB(filehash)
	log.Println(22222)
	if err != nil {
		c.JSON(http.StatusInternalServerError, "the filehash of file is not exist")
		return
	}

	data, err := json.Marshal(fMetaResult)
	if err != nil{
		log.Printf("Failed to tran result json, err: %s", err.Error())
		c.JSON(http.StatusInternalServerError, "Failed to tran result json")
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

func FileQueryHandler(c *gin.Context) {

	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	//fileMetas := meta.GetListFileMetas(limitCnt)
	//fileMetas, err := meta.GetListFileMetasDB(limitCnt)
	username := c.Request.FormValue("username")
	userFiles, err := dao.QueryFileMetas(username, limitCnt)
	if err != nil{

		c.JSON(http.StatusInternalServerError, "query userfile error!")
		return
	}
	data, err := json.Marshal(userFiles)
	if err != nil {

		c.JSON(http.StatusInternalServerError, "Failed to tran result json")
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}

//DownloadHandler：根据filehash下载文件
func DownloadHandler(c *gin.Context){

	fsha1 := c.Request.FormValue("filehash")
	fm,err := meta.GetFileMetaDB(fsha1)
	if err != nil{
		fmt.Printf("No This File, err: %s", err.Error())
		c.JSON(http.StatusInternalServerError, "Get File Meta error")
		return
	}

	f, err := os.Open(fm.Location)
	if err != nil{
		fmt.Printf("Failed to open file, err: %s", err.Error())
		c.JSON(http.StatusInternalServerError, "Failed to open file")
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Printf("Failed to read file, err: %s", err.Error())
		c.JSON(http.StatusInternalServerError, "Failed to open file")
		return
	}
	c.Data(http.StatusOK, "application/octect-stream", data)

}

// FileMetaUpdateHandler ： 更新元信息接口(重命名)
func FileMetaUpdateHandler(c *gin.Context) {
	opType := c.Request.FormValue("op")
	fileSha1 := c.Request.FormValue("filehash")
	newFileName := c.Request.FormValue("filename")

	if opType != "0" {
		c.JSON(http.StatusForbidden, "opTye error")
		return
	}

	curFileMeta, err := meta.GetFileMetaDB(fileSha1)
	if err != nil{
		fmt.Printf("No This File, err: %s", err.Error())
		c.JSON(http.StatusInternalServerError, "No This File")
		return
	}

	curFileMeta.FileName = newFileName
	ret := meta.UpdateFileMetaDB(*curFileMeta)
	if ret == true{
		data, err := json.Marshal(curFileMeta)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Json Trans Error")
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	} else {
		data, err := json.Marshal("update filed!")
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Json Trans Error")
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	}

}

// FileDeleteHandler : 删除文件及元信息
func FileDeleteHandler(c *gin.Context) {
	fileSha1 := c.Request.FormValue("filehash")
	fMeta,err := meta.GetFileMetaDB(fileSha1)
	if err != nil {
		fmt.Printf("No This File, err: %s", err.Error())
		c.JSON(http.StatusInternalServerError, "No This File, err")
		return
	}
	// 删除文件
	os.Remove(fMeta.Location)
	// 删除文件元信息
	res, err := meta.RemoveFileMetaDB(fileSha1)
	if res != true{
		c.JSON(http.StatusInternalServerError, "Remove This File, err")
	}

	c.JSON(http.StatusOK, "Remove Ok")
}

//TryFastUploadHandler:尝试秒传接口
func TryFastUploadHandler(c *gin.Context) {

	//解析参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize, _ := strconv.Atoi(c.Request.FormValue("filesize"))

	//从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, "get data error")
		return
	}

	//查询不到记录则返回查询失败
	if fileMeta == nil {
		resp := utils.RespMsg{
			Code: -1,
			Msg: "秒传失败，请访问普通接口！",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}

	//查询到就将文件信息写入用户文件表，返回成功即可
	suc := dao.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))

	if suc {
		resp := utils.RespMsg{
			Code: 0,
			Msg: "秒传成功！",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	} else {
		resp := utils.RespMsg{
			Code: -2,
			Msg: "秒传失败，请稍后重试！",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}

}


// DownloadURLHandler : 生成文件的下载地址
func DownloadURLHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")

	//从文件表中查找
	row, _ := dao.GetFileMeta(filehash)

	//Todo:判断文件是存在OSS中还是Ceph中

	signURL := oss.DownloadURL(row.FileAddr.String)
	c.JSON(http.StatusOK, signURL)

}