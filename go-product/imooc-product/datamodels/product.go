package datamodels

type Product struct {
	ID           int64  `json:"id" sql:"id" imooc:"id"`
	ProductName  string `json:"productName" sql:"productName" imooc:"productName"`
	ProductNum   int64  `json:"productNum" sql:"productNum" imooc:"productNum"`
	ProductImage string `json:"productImage" sql:"productImage" imooc:"productImage"`
	ProductUrl   string `json:"productUrl" sql:"productUrl" imooc:"productUrl"`
}

const (
	OrderWait = iota
	OrderSuccess
	OrderFailed
)
