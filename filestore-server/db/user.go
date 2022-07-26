package db

import (
	mydb "filestore-server/db/mysql"
	"log"
)

type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// UserSignUp: 注册用户
func UserSignUp(username string, password string) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"INSERT IGNORE INTO `tbl_user`(`user_name`, `user_pwd`) VALUES(?, ?)",
	)
	if err != nil {
		log.Printf("Failed to insert, err: %s\n", err.Error())
		return false, err
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, password)
	if err != nil {
		log.Printf("Failed to insert, err: %s\n", err.Error())
		return false, err
	}

	if cnt, err := ret.RowsAffected(); err != nil {
		log.Printf("Failed to insert, err: %s\n", err.Error())
		return false, err
	} else if cnt <= 0 {
		log.Printf("duplicated user: %s\n", username)
		return false, nil
	}

	return true, nil
}

// UserSignIn：用户登录（校验数据库中的用户名和密码是否正确）
func UserSignIn(username string, encPassword string) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT * FROM `tbl_user` WHERE `user_name` = ? LIMIT 1",
	)
	if err != nil {
		log.Printf("Failed to select user(prepare), err: %s\n", err.Error())
		return false, err
	}
	defer stmt.Close()

	// TODO: 改成 QueryRow
	rows, err := stmt.Query(username)
	if err != nil {
		log.Printf("Failed to query user, err: %s\n", err.Error())
		return false, err
	} else if rows == nil {
		log.Printf("Username %s not found\n", username)
		return false, nil
	}
	defer rows.Close()

	records, err := mydb.ParseRows(rows)
	if err != nil {
		log.Printf("Failed to save sql.Rows into map, err: %s\n", err.Error())
		return false, err
	} else if len(records) == 0 || string(records[0]["user_pwd"].([]byte)) != encPassword {
		return false, nil
	}

	return true, nil
}

// UpdateToken：刷新用户 token
func UpdateToken(username string, token string) (bool, error) {
	stmt, err := mydb.DBConn().Prepare(
		// replace into: 如果发现表中已经有此行数据（根据主键或者唯一索引判断）则先删除此行数据，然后插入新的数据。否则，直接插入新数据。
		"REPLACE INTO `tbl_user_token`(`user_name`, `user_token`) VALUES(?, ?)",
	)
	if err != nil {
		log.Printf("Failed to update token(prepare), err: %s\n", err.Error())
		return false, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		log.Printf("Failed to update token(stmt), err: %s\n", err.Error())
		return false, err
	}
	return true, nil
}

// GetUserInfo: 获取用户信息
func GetUserInfo(username string) (*User, error) {
	stmt, err := mydb.DBConn().Prepare(
		"SELECT `user_name`, `signup_at` FROM `tbl_user` WHERE `user_name` = ? LIMIT 1",
	)
	if err != nil {
		log.Printf("Failed to select user(prepare), err: %s\n", err.Error())
		return nil, err
	}
	defer stmt.Close()

	user := &User{}
	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		log.Printf("Failed to query user, err: %s\n", err.Error())
		return nil, err
	}

	return user, nil
}
