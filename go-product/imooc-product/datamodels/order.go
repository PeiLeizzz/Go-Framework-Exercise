package datamodels

type Order struct {
	ID          int64 `sql:"id"`
	UserID      int64 `sql:"userID"`
	ProductID   int64 `sql:"productID"`
	OrderStatus int64 `sql:"orderStatus"`
}
