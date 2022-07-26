package repositories

import (
	"database/sql"
	"go-product/imooc-product/common"
	"go-product/imooc-product/datamodels"
	"strconv"
)

type IProduct interface {
	// 连接数据库
	Conn() error
	Insert(*datamodels.Product) (int64, error)
	Delete(int64) bool
	Update(*datamodels.Product) error
	SelectByKey(int64) (*datamodels.Product, error)
	SelectAll() ([]*datamodels.Product, error)
	SubProductNum(int64) error
}

type ProductManager struct {
	table     string
	mysqlConn *sql.DB
}

var _ IProduct = (*ProductManager)(nil)

func NewProductManager(table string, db *sql.DB) IProduct {
	return &ProductManager{
		table:     table,
		mysqlConn: db,
	}
}

func (p *ProductManager) Conn() error {
	if p.mysqlConn == nil {
		mysql, err := common.NewMysqlConn()
		if err != nil {
			return err
		}
		p.mysqlConn = mysql
	}
	if p.table == "" {
		p.table = "product"
	}
	return nil
}

func (p *ProductManager) Insert(product *datamodels.Product) (int64, error) {
	// 判断连接是否正常
	if err := p.Conn(); err != nil {
		return 0, err
	}

	sql := "INSERT product SET productName=?, productNum=?, productImage=?, productUrl=?"

	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return 0, err
	}

	result, err := stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (p *ProductManager) Delete(productID int64) bool {
	if err := p.Conn(); err != nil {
		return false
	}

	sql := "DELETE FROM product WHERE ID=?"

	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return false
	}

	_, err = stmt.Exec(productID)
	if err != nil {
		return false
	}

	return true
}

func (p *ProductManager) Update(product *datamodels.Product) error {
	if err := p.Conn(); err != nil {
		return err
	}

	sql := "UPDATE product SET productName=?, productNum=?, productImage=?, productUrl=? WHERE id=" +
		strconv.FormatInt(product.ID, 10)

	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(product.ProductName, product.ProductNum, product.ProductImage, product.ProductUrl)

	return err
}

func (p *ProductManager) SelectByKey(productID int64) (*datamodels.Product, error) {
	product := &datamodels.Product{}
	if err := p.Conn(); err != nil {
		return product, err
	}

	sql := "SELECT * from " + p.table + " WHERE id = " + strconv.FormatInt(productID, 10)

	row, err := p.mysqlConn.Query(sql)
	defer row.Close()
	if err != nil {
		return product, err
	}

	result := common.GetResultRow(row)
	if len(result) == 0 {
		return product, nil
	}

	common.DataToStructByTagSql(result, product)
	return product, nil
}

func (p *ProductManager) SelectAll() ([]*datamodels.Product, error) {
	if err := p.Conn(); err != nil {
		return nil, err
	}

	sql := "SELECT * from " + p.table

	rows, err := p.mysqlConn.Query(sql)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	results := common.GetResultRows(rows)
	if len(results) == 0 {
		return nil, nil
	}

	products := []*datamodels.Product{}
	for _, v := range results {
		product := &datamodels.Product{}
		common.DataToStructByTagSql(v, product)
		products = append(products, product)
	}
	return products, nil
}

func (p *ProductManager) SubProductNum(productID int64) error {
	if err := p.Conn(); err != nil {
		return err
	}

	sql := "UPDATE " + p.table + " SET " + "productNum=productNum-1 WHERE ID = " + strconv.FormatInt(productID, 10)

	stmt, err := p.mysqlConn.Prepare(sql)
	if err != nil {
		return err
	}

	_, err = stmt.Exec()
	return err
}
