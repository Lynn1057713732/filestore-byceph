package handler

import (
	"encoding/json"
	"filestore-byceph/utils"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"filestore-byceph/meta"
)


//UploadHandler:处理用户注册请求
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//返回上传html页面
		data, err := ioutil.ReadFile("./static/view/home.html")
		if err != nil {
			io.WriteString(w, "internet server error!")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		//接收文件流及存储到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data, err: %s", err.Error())
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "/Users/yangfengming/opt/"+head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil{
			fmt.Printf("Failed to created file, err: %s", err.Error())
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to created file, err: %s", err.Error())
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = utils.FileSha1(newFile)
		//meta.UpdateFileMeta(fileMeta)
		_ = meta.CreateFileMetaDB(fileMeta)
		//使用重定向
		http.Redirect(w, r, "/file/upload/success", http.StatusFound)

	}

}


//UploadSuccessHandler：上传已完成
func UploadSuccessHandler(w http.ResponseWriter, r *http.Request) {

	io.WriteString(w, "Upload finished!")
}


//GetFileMetaHandler：获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	filehash := r.Form["filehash"][0]

	//fMetaResult := meta.GetFileMeta(filehash)
	fMetaResult, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(fMetaResult)
	if err != nil{
		fmt.Printf("Failed to tran result json, err: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	//fileMetas := meta.GetListFileMetas(limitCnt)
	fileMetas, err := meta.GetListFileMetasDB(limitCnt)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fileMetas)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

//DownloadHandler：根据filehash下载文件
func DownloadHandler(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fm,err := meta.GetFileMetaDB(fsha1)
	if err != nil{
		fmt.Printf("No This File, err: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	f, err := os.Open(fm.Location)
	if err != nil{
		fmt.Printf("Failed to open file, err: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Printf("Failed to read file, err: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment;filename=\""+fm.FileName+"\"")
	w.Write(data)

}

// FileMetaUpdateHandler ： 更新元信息接口(重命名)
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "PATCH" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta, err := meta.GetFileMetaDB(fileSha1)
	if err != nil{
		fmt.Printf("No This File, err: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	curFileMeta.FileName = newFileName
	ret := meta.UpdateFileMetaDB(*curFileMeta)
	if ret == true{
		data, err := json.Marshal(curFileMeta)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	} else {
		data, err := json.Marshal("update filed!")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
	}

}

// FileDeleteHandler : 删除文件及元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fMeta,err := meta.GetFileMetaDB(fileSha1)
	if err != nil {
		fmt.Printf("No This File, err: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 删除文件
	os.Remove(fMeta.Location)
	// 删除文件元信息
	res, err := meta.RemoveFileMetaDB(fileSha1)
	if res != true{
		w.WriteHeader(http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
}
