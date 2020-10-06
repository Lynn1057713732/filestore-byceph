package db

import (
	"database/sql"
	mydb "filestore-byceph/db/mysql"
	"fmt"
)

// OnFileUploadFinished : 文件上传完成，保存meta
func OnFileUploadFinished(filehash string, filename string,
	filesize int64, fileaddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_file (`file_sha1`,`file_name`,`file_size`," +
			"`file_addr`,`status`) values (?,?,?,?,1)")
	if err != nil {
		fmt.Println("Failed to prepare statement, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		if rf <= 0 {
			fmt.Printf("File with hash:%s has been uploaded before", filehash)
		}
		return true
	}
	return false
}

// TableFile : 文件表结构体
type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}


// GetFileMeta : 从mysql获取文件元信息
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where file_sha1=? and status=1 limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	//确保QueryRow之后调用Scan方法，否则持有的数据库链接不会被释放
	err = stmt.QueryRow(filehash).Scan(
		&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
	if err != nil {
		if err == sql.ErrNoRows {
			// 查不到对应记录， 返回参数及错误均为nil
			return nil, nil
		} else {
			fmt.Println(err.Error())
			return nil, err
		}
	}
	return &tfile, nil
}

// GetFileMetaList : 从mysql批量获取文件元信息
func GetFileMetaList(limit int) ([]TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"select file_sha1,file_addr,file_name,file_size from tbl_file " +
			"where status=1 limit ?")
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	cloumns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(cloumns))
	var tfiles []TableFile
	for i := 0; i < len(values) && rows.Next(); i++ {
		tfile := TableFile{}
		err = rows.Scan(&tfile.FileHash, &tfile.FileAddr,
			&tfile.FileName, &tfile.FileSize)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		tfiles = append(tfiles, tfile)
	}
	fmt.Println(len(tfiles))
	return tfiles, nil
}

func UpdateFileMetaByFileHash(filehash string, filename string)(bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"Update tbl_file SET file_name=? where tbl_file.file_sha1 = ?")
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(filename, filehash)
	if err != nil {
		fmt.Printf("update faied, error:[%v]", err.Error())
		return false, err
	}
	n, err := result.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return false, err
	}
	fmt.Printf("update success, affected rows:%d\n", n)

	return true, nil

}

func DeleteFileMetaByFileHash(filehash string)(bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"Update tbl_file SET status=? where tbl_file.file_sha1 = ?")
	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(2, filehash)
	if err != nil {
		fmt.Printf("dalete faied, error:[%v]", err.Error())
		return false, err
	}
	n, err := result.RowsAffected() // 操作影响的行数
	if err != nil {
		fmt.Printf("get RowsAffected failed, err:%v\n", err)
		return false, err
	}
	fmt.Printf("datele success, affected rows:%d\n", n)

	return true, nil

}