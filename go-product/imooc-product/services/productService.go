package services

import (
	"go-product/imooc-product/datamodels"
	"go-product/imooc-product/repositories"
)

type IProductService interface {
	GetProductByID(int64) (*datamodels.Product, error)
	GetAllProduct() ([]*datamodels.Product, error)
	DeleteProductByID(int64) bool
	InsertProduct(*datamodels.Product) (int64, error)
	UpdateProduct(*datamodels.Product) error
	SubNumberOne(int64) error
}

type ProductService struct {
	productRepository repositories.IProduct
}

var _ IProductService = (*ProductService)(nil)

func NewProductService(repository repositories.IProduct) IProductService {
	return &ProductService{
		productRepository: repository,
	}
}

func (p *ProductService) GetProductByID(productID int64) (*datamodels.Product, error) {
	return p.productRepository.SelectByKey(productID)
}

func (p *ProductService) GetAllProduct() ([]*datamodels.Product, error) {
	return p.productRepository.SelectAll()
}

func (p *ProductService) DeleteProductByID(productID int64) bool {
	return p.productRepository.Delete(productID)
}

func (p *ProductService) InsertProduct(product *datamodels.Product) (int64, error) {
	return p.productRepository.Insert(product)
}

func (p *ProductService) UpdateProduct(product *datamodels.Product) error {
	return p.productRepository.Update(product)
}

func (p *ProductService) SubNumberOne(productID int64) error {
	return p.productRepository.SubProductNum(productID)
}
