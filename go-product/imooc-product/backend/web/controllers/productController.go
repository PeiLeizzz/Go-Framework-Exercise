package controllers

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go-product/imooc-product/common"
	"go-product/imooc-product/datamodels"
	"go-product/imooc-product/services"
	"strconv"
)

type ProductController struct {
	Ctx            iris.Context
	ProductService services.IProductService
}

func (p *ProductController) failOnErr(err error) mvc.View {
	p.Ctx.Application().Logger().Debug(err)
	return mvc.View{
		Name: "shared/error.html",
	}
}

func (p *ProductController) GetAll() mvc.View {
	products, _ := p.ProductService.GetAllProduct()
	return mvc.View{
		Name: "product/view.html",
		Data: iris.Map{
			"products": products,
		},
	}
}

func (p *ProductController) GetManager() mvc.View {
	idString := p.Ctx.URLParam("id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		return p.failOnErr(err)
	}

	product, err := p.ProductService.GetProductByID(id)
	if err != nil {
		return p.failOnErr(err)
	}

	return mvc.View{
		Name: "product/manager.html",
		Data: iris.Map{
			"product": product,
		},
	}
}

func (p *ProductController) GetAdd() mvc.View {
	return mvc.View{
		Name: "product/add.html",
	}
}

func (p *ProductController) PostUpdate() {
	product := &datamodels.Product{}
	p.Ctx.Request().ParseForm()
	dec := common.NewDecoder(&common.DecoderOptions{TagName: "imooc"})

	if err := dec.Decode(p.Ctx.Request().Form, product); err != nil {
		p.failOnErr(err)
	}

	if err := p.ProductService.UpdateProduct(product); err != nil {
		p.failOnErr(err)
	}
	p.Ctx.Redirect("/product/all")
}

func (p *ProductController) PostAdd() {
	product := &datamodels.Product{}
	p.Ctx.Request().ParseForm()
	dec := common.NewDecoder(&common.DecoderOptions{TagName: "imooc"})

	if err := dec.Decode(p.Ctx.Request().Form, product); err != nil {
		p.failOnErr(err)
	}

	if _, err := p.ProductService.InsertProduct(product); err != nil {
		p.failOnErr(err)
	}
	p.Ctx.Redirect("/product/all")
}

func (p *ProductController) GetDelete() {
	idString := p.Ctx.URLParam("id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil {
		p.failOnErr(err)
	}

	if ok := p.ProductService.DeleteProductByID(id); ok {
		p.Ctx.Application().Logger().Debug("success, id: " + idString)
	} else {
		p.failOnErr(err)
	}

	p.Ctx.Redirect("/product/all")
}
