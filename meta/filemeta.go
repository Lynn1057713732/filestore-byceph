package meta

import (
	mydb "filestore-byceph/db"
	"fmt"
)

//FileMeta：文件元信息结构体
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

//var fileMetas map[string]FileMeta
//
//func init() {
//	fileMetas = make(map[string]FileMeta)
//}


// CreateFileMetaDB : 新增文件元信息到mysql中
func CreateFileMetaDB(fmeta FileMeta) bool{
	return mydb.OnFileUploadFinished(
		fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
}

// UpdateFileMetaDB : 更新文件元信息到mysql中
func UpdateFileMetaDB(fmeta FileMeta) bool{
	ret, err := mydb.UpdateFileMetaByFileHash(fmeta.FileSha1, fmeta.FileName)
	if err != nil{
		fmt.Printf("update file meta error: %s", err.Error())

	}
	return ret
}


// GetFileMetaDB : 从mysql获取文件元信息
func GetFileMetaDB(fileSha1 string) (*FileMeta, error) {
	tfile, err := mydb.GetFileMeta(fileSha1)
	if tfile == nil || err != nil {
		return nil, err
	}
	fmeta := FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return &fmeta, nil
}

// 删除元信息
func RemoveFileMetaDB(fileSha1 string) (bool, error){
	res, err := mydb.DeleteFileMetaByFileHash(fileSha1)
	return res, err

}


// GetListFileMetasDB : 批量从mysql获取文件元信息
func GetListFileMetasDB(limit int) ([]FileMeta, error) {
	tfiles, err := mydb.GetFileMetaList(limit)
	if err != nil {
		return make([]FileMeta, 0), err
	}

	tfilesm := make([]FileMeta, len(tfiles))
	for i := 0; i < len(tfilesm); i++ {
		tfilesm[i] = FileMeta{
			FileSha1: tfiles[i].FileHash,
			FileName: tfiles[i].FileName.String,
			FileSize: tfiles[i].FileSize.Int64,
			Location: tfiles[i].FileAddr.String,
		}
	}
	return tfilesm, nil
}