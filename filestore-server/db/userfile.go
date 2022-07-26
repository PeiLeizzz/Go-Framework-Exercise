package db

import (
	mydb "filestore-server/db/mysql"
	"log"
	"time"
)

type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// OnUserFileUploadFinished：添加用户文件表
func OnUserFileUploadFinished(username string, filehash string, filename string, filesize int64) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"INSERT IGNORE INTO `tbl_user_file`(`user_name`, `file_sha1`, `file_name`, " +
			"`file_size`, `upload_at`) VALUES(?, ?, ?, ?, ?)",
	)
	if err != nil {
		log.Printf("Failed to insert userfile(prepare), err: %s\n", err.Error())
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, filehash, filename, filesize, time.Now())
	if err != nil {
		log.Printf("Failed to insert userfile(exec), err: %s\n", err.Error())
		return false, err
	}
	// TODO:
	// username + file_hash 都相同，其他不同的话，也会失败（unique key 相同，插入会被 ignore）
	// 这里返回的 result.RowsEffect() == 0
	// 先忽略这个问题

	return true, nil
}

// QueryUserFileMetas: 批量获取用户文件信息
func QueryUserFileMetas(username string, limit int) ([]*UserFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT `file_sha1`, `file_name`, `file_size`, `upload_at`, `last_update` FROM " +
			"`tbl_user_file` WHERE `user_name` = ? LIMIT ?",
	)
	if err != nil {
		log.Printf("Failed to select from userfile(prepare), err: %s\n", err.Error())
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(username, limit)
	if err != nil {
		log.Printf("Failed to query userfile, err: %s\n", err.Error())
		return nil, err
	}
	defer rows.Close()

	var userFiles []*UserFile
	for rows.Next() {
		ufile := &UserFile{}
		err = rows.Scan(&ufile.FileHash, &ufile.FileName, &ufile.FileSize, &ufile.UploadAt, &ufile.LastUpdated)
		if err != nil {
			log.Printf("Failed to query userfile row, err: %s\n", err.Error())
			return nil, err
		}
		userFiles = append(userFiles, ufile)
	}

	return userFiles, nil
}
