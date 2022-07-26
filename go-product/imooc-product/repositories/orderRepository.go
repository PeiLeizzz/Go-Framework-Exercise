package repositories

import (
	"database/sql"
	"go-product/imooc-product/common"
	"go-product/imooc-product/datamodels"
	"strconv"
)

type IOrderRepository interface {
	Conn() error
	Insert(*datamodels.Order) (int64, error)
	Delete(int64) bool
	Update(*datamodels.Order) error
	SelectByKey(int64) (*datamodels.Order, error)
	SelectAll() ([]*datamodels.Order, error)
	SelectAllWithInfo() (map[int]map[string]string, error)
}

type OrderManager struct {
	table     string
	mysqlConn *sql.DB
}

var _ IOrderRepository = (*OrderManager)(nil)

func NewOrderManager(table string, db *sql.DB) IOrderRepository {
	return &OrderManager{
		table:     table,
		mysqlConn: db,
	}
}

func (o *OrderManager) Conn() error {
	if o.mysqlConn == nil {
		mysql, err := common.NewMysqlConn()
		if err != nil {
			return err
		}
		o.mysqlConn = mysql
	}
	if o.table == "" {
		o.table = "order"
	}
	return nil
}

func (o *OrderManager) Insert(order *datamodels.Order) (int64, error) {
	if err := o.Conn(); err != nil {
		return 0, err
	}

	sql := "INSERT `order` SET userID = ?, productID = ?, orderStatus = ?"

	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return 0, err
	}

	result, err := stmt.Exec(order.UserID, order.ProductID, order.OrderStatus)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (o *OrderManager) Delete(productID int64) bool {
	if err := o.Conn(); err != nil {
		return false
	}

	sql := "DELETE FROM " + o.table + " WHERE id = ?"

	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return false
	}

	_, err = stmt.Exec(productID)
	if err != nil {
		return false
	}

	return true
}

func (o *OrderManager) Update(order *datamodels.Order) error {
	if err := o.Conn(); err != nil {
		return err
	}

	sql := "UPDATE " + o.table + " SET userID = ?, productID = ?, orderStatus = ? WHERE id = " +
		strconv.FormatInt(order.ID, 10)

	stmt, err := o.mysqlConn.Prepare(sql)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(order.UserID, order.ProductID, order.OrderStatus)

	return err
}

func (o *OrderManager) SelectByKey(orderID int64) (*datamodels.Order, error) {
	order := &datamodels.Order{}
	if err := o.Conn(); err != nil {
		return order, err
	}

	sql := "SELECT * FROM " + o.table + " WHERE id = " + strconv.FormatInt(orderID, 10)

	row, err := o.mysqlConn.Query(sql)
	defer row.Close()
	if err != nil {
		return order, err
	}

	result := common.GetResultRow(row)
	if len(result) == 0 {
		return order, nil
	}

	common.DataToStructByTagSql(result, order)
	return order, nil
}

func (o *OrderManager) SelectAll() ([]*datamodels.Order, error) {
	if err := o.Conn(); err != nil {
		return nil, err
	}

	sql := "SELECT * FROM " + o.table

	rows, err := o.mysqlConn.Query(sql)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	result := common.GetResultRows(rows)
	if len(result) == 0 {
		return nil, err
	}

	orders := []*datamodels.Order{}
	for _, v := range result {
		order := &datamodels.Order{}
		common.DataToStructByTagSql(v, order)
		orders = append(orders, order)
	}

	return orders, nil
}

// 查询和订单有关的信息（订单号、商品名、订单状态）
func (o *OrderManager) SelectAllWithInfo() (map[int]map[string]string, error) {
	if err := o.Conn(); err != nil {
		return nil, err
	}

	sql := "SELECT o.ID, p.productName, o.orderStatus FROM imooc.order AS o LEFT JOIN imooc.product AS p ON o.productID = p.id"

	rows, err := o.mysqlConn.Query(sql)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	return common.GetResultRows(rows), nil
}
