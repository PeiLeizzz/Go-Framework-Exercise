package repositories

import (
	"database/sql"
	"errors"
	"go-product/imooc-product/common"
	"go-product/imooc-product/datamodels"
	"strconv"
)

type IUserRepository interface {
	Conn() error
	SelectByUserName(string) (*datamodels.User, error)
	SelectByID(int64) (*datamodels.User, error)
	Insert(*datamodels.User) (int64, error)
}

type UserManager struct {
	table     string
	mysqlConn *sql.DB
}

var _ IUserRepository = (*UserManager)(nil)

func NewUserManager(table string, db *sql.DB) IUserRepository {
	return &UserManager{
		table:     table,
		mysqlConn: db,
	}
}

func (u *UserManager) Conn() error {
	if u.mysqlConn == nil {
		mysql, err := common.NewMysqlConn()
		if err != nil {
			return err
		}
		u.mysqlConn = mysql
	}
	if u.table == "" {
		u.table = "user"
	}
	return nil
}

func (u *UserManager) SelectByUserName(userName string) (*datamodels.User, error) {
	user := &datamodels.User{}
	if userName == "" {
		return user, errors.New("userName cannot be empty!")
	}

	if err := u.Conn(); err != nil {
		return user, err
	}

	sql := "SELECT * FROM " + u.table + " WHERE userName = ?"

	row, err := u.mysqlConn.Query(sql, userName)
	defer row.Close()
	if err != nil {
		return user, err
	}

	result := common.GetResultRow(row)
	if len(result) == 0 {
		return user, errors.New("user not existed!")
	}

	common.DataToStructByTagSql(result, user)
	return user, nil
}

func (u *UserManager) Insert(user *datamodels.User) (int64, error) {
	if err := u.Conn(); err != nil {
		return 0, err
	}

	sql := "INSERT " + u.table + " SET nickName = ?, userName = ?, password = ?"

	stmt, err := u.mysqlConn.Prepare(sql)
	if err != nil {
		return 0, err
	}

	result, err := stmt.Exec(user.NickName, user.UserName, user.HashPassword)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (u *UserManager) SelectByID(userID int64) (*datamodels.User, error) {
	user := &datamodels.User{}
	if err := u.Conn(); err != nil {
		return user, err
	}

	sql := "SELECT * FROM " + u.table + " WHERE ID = " + strconv.FormatInt(userID, 10)

	row, err := u.mysqlConn.Query(sql)
	defer row.Close()
	if err != nil {
		return user, err
	}

	result := common.GetResultRow(row)
	if len(result) == 0 {
		return user, errors.New("user not existed!")
	}

	common.DataToStructByTagSql(result, user)
	return user, nil
}
