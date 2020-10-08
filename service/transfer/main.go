package transfer

import (
	"os"
	"log"
	"bufio"
	"encoding/json"

	"filestore-byceph/mq"
	"filestore-byceph/store/oss"
	dao "filestore-byceph/db"
)

//ProcessTransfer:处理文件转移的真正逻辑
func ProcessTransfer(msg []byte) bool {
	//解析msg
	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	//根据临时存储文件路径，创建文件句柄
	fin, err := os.Open(pubData.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	//通过文件句柄将文件内容读出来并且上传到OSS
	err = oss.Bucket().PutObject(
		pubData.DestLocation,
		bufio.NewReader(fin))
	if err != nil {
		log.Println(err.Error())
		return false
	}

	//更新文件的存储路径到文件表
	resp, err := dao.UpdateFileLocation(
		pubData.FileHash,
		pubData.DestLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}
	if !resp.Suc {
		log.Println("更新数据库异常，请检查:" + pubData.FileHash)
		return false
	}
	return true
}
