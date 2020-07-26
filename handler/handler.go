package handler

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

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
		fmt.Println("aaaa")
		if err != nil {
			fmt.Printf("Failed to get data, err: %s", err.Error())
			return
		}
		defer file.Close()

		newFile, err := os.Create("/Users/yangfengming/opt/"+head.Filename)
		if err != nil{
			fmt.Printf("Failed to created file, err: %s", err.Error())
			return
		}
		defer newFile.Close()

		_, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to created file, err: %s", err.Error())
			return
		}

		http.Redirect(w, r, "/file/upload/success", http.StatusFound)

	}

}


//UploadSuccessHandler：上传已完成
func UploadSuccessHandler(w http.ResponseWriter, r *http.Request) {

	io.WriteString(w, "Upload finished!")
}



