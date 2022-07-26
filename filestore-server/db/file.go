package db

import (
	"database/sql"
	mydb "filestore-server/db/mysql"
	"log"
)

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// OnFileUploadFinished: 文件上传完成，保存元信息
func OnFileUploadFinished(filehash string, filename string, filesize int64, fileaddr string) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"INSERT IGNORE INTO `tbl_file`(`file_sha1`, `file_name`, `file_size`," +
			"`file_addr`, `status`) VALUES(?, ?, ?, ?, 1)",
	)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return false, err
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	if cnt, err := ret.RowsAffected(); err != nil {
		log.Println(err.Error())
		return false, err
	} else if cnt <= 0 {
		log.Printf("File with hash: %s has been uploaded before\n", filehash)
	}

	return true, nil
}

// UpdateFileMeta: 更新文件元信息
func UpdateFileMeta(filehash string, filename string, filesize int64, fileaddr string) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"UPDATE `tbl_file` SET `file_name` = ?, `file_size` = ?," +
			"`file_addr` = ? WHERE `file_sha1` = '" + filehash + `'`,
	)
	if err != nil {
		log.Printf("Failed to prepare statement, err: %s\n", err.Error())
		return false, err
	}
	defer stmt.Close()

	ret, err := stmt.Exec(filename, filesize, fileaddr)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	if cnt, err := ret.RowsAffected(); err != nil {
		log.Println(err.Error())
		return false, err
	} else if cnt <= 0 {
		log.Printf("No changes: %s\n", filehash)
		return true, nil
	}

	return true, nil
}

// GetFileMeta: 根据 filehash 从数据库中获取文件元信息
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT `file_sha1`, `file_addr`, `file_name`, `file_size` FROM `tbl_file` WHERE " +
			"`file_sha1` = ? AND `status` = 1 LIMIT 1",
	)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := &TableFile{}
	err = stmt.QueryRow(filehash).Scan(&tfile.FileHash, &tfile.FileAddr, &tfile.FileName, &tfile.FileSize)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	return tfile, nil
}

func UpdateFileLocation(filehash string, fileaddr string) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"UPDATE `tbl_file` SET `file_addr` = ? WHERE `file_sha1` = ? LIMIT 1",
	)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		log.Println(err.Error())
		return false, err
	}

	if rf, err := ret.RowsAffected(); err != nil {
		log.Println(err.Error())
		return false, err
	} else if rf <= 0 {
		log.Println("Failed to update file location, filehash: %s, err: %s\n", filehash, err.Error())
		return false, nil
	}

	return true, nil
}
