package handler

import (
	"encoding/json"
	"filestore-byceph/store/ceph"
	"filestore-byceph/utils"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	cfg "filestore-byceph/config"
	"filestore-byceph/meta"
	dao "filestore-byceph/db"
)


//UploadHandler:处理文件上传
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


		//同时将文件写入ceph存储中
		newFile.Seek(0, 0)
		data, _ := ioutil.ReadAll(newFile)
		cephPath := cfg.CephRootDir + fileMeta.FileSha1
		_ = ceph.PutObject("userfile", cephPath, data)
		fileMeta.Location = cephPath

		//meta.UpdateFileMeta(fileMeta)
		_ = meta.CreateFileMetaDB(fileMeta)

		//更新用户文件表
		r.ParseForm()
		username := r.Form.Get("username")
		fmt.Printf("username: %s", username)

		suc := dao.OnUserFileUploadFinished(username, fileMeta.FileSha1,
			fileMeta.FileName, fileMeta.FileSize)
		fmt.Printf("%b", suc)
		if suc {
			//使用重定向
			http.Redirect(w, r, "/file/upload/success", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed!"))

		}


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
	//fileMetas, err := meta.GetListFileMetasDB(limitCnt)
	username := r.Form.Get("username")
	userFiles, err := dao.QueryFileMetas(username, limitCnt)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(userFiles)
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

//TryFastUploadHandler:尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	//解析参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	//从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//查询不到记录则返回查询失败
	if fileMeta == nil {
		resp := utils.RespMsg{
			Code: -1,
			Msg: "秒传失败，请访问普通接口！",
		}
		w.Write(resp.JSONBytes())
		return
	}

	//查询到就将文件信息写入用户文件表，返回成功即可
	suc := dao.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))

	if suc {
		resp := utils.RespMsg{
			Code: 0,
			Msg: "秒传成功！",
		}
		w.Write(resp.JSONBytes())
		return
	} else {
		resp := utils.RespMsg{
			Code: -2,
			Msg: "秒传失败，请稍后重试！",
		}
		w.Write(resp.JSONBytes())
		return
	}


}