package controllers

import (
	"encoding/json"
	"errors"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go-product/RabbitMQ"
	"go-product/imooc-product/datamodels"
	"go-product/imooc-product/services"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
)

type ProductController struct {
	Ctx            iris.Context
	ProductService services.IProductService
	OrderService   services.IOrderService
	RabbitMQ       *RabbitMQ.RabbitMQSimple
}

var (
	htmlOutPath  = "./frontend/web/htmlProductShow/" // 生成的 html 保存目录
	templatePath = "./frontend/web/views/template/"  // 静态文件模板目录
)

func (p *ProductController) failOnErr(err error) mvc.View {
	p.Ctx.Application().Logger().Error(err)
	return mvc.View{
		Name: "shared/error.html",
	}
}

func (p *ProductController) GetGenerateHtml() {
	// 1. 获取模板文件地址
	contentTmp, err := template.ParseFiles(filepath.Join(templatePath, "product.html"))
	if err != nil {
		p.failOnErr(err)
		return
	}

	// 2. 获取 html 生成路径
	fileName := filepath.Join(htmlOutPath, "htmlProduct.html")

	// 3. 获取模板渲染数据
	productID, err := strconv.ParseInt(p.Ctx.URLParam("productID"), 10, 64)
	if err != nil {
		p.failOnErr(err)
	}
	product, err := p.ProductService.GetProductByID(productID)
	if err != nil {
		p.failOnErr(err)
	}

	// 4. 生成静态文件
	generateStaticHtml(p.Ctx, contentTmp, fileName, product)
}

func generateStaticHtml(ctx iris.Context, template *template.Template, fileName string, product *datamodels.Product) {
	// 1. 判断静态文件是否存在，如果存在则删除
	if exist(fileName) {
		err := os.Remove(fileName)
		if err != nil {
			ctx.Application().Logger().Error(err)
			return
		}
	}

	// 2. 生成静态文件
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		ctx.Application().Logger().Error(err)
		return
	}
	defer file.Close()

	template.Execute(file, &product)
}

// 判断文件是否已经生成
func exist(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil || os.IsExist(err)
}

func (p *ProductController) GetDetail() mvc.View {
	productID, err := strconv.ParseInt(p.Ctx.URLParam("productID"), 10, 64)
	if err != nil {
		return p.failOnErr(err)
	}

	product, err := p.ProductService.GetProductByID(productID)
	if err != nil {
		return p.failOnErr(err)
	}

	return mvc.View{
		Layout: "shared/productLayout.html",
		Name:   "product/view.html",
		Data: iris.Map{
			"product": product,
		},
	}
}

func (p *ProductController) GetOrder() []byte {
	productID, err := strconv.ParseInt(p.Ctx.URLParam("productID"), 10, 64)
	if err != nil {
		p.failOnErr(err)
	}

	userIDString := p.Ctx.GetCookie("uid")
	userID, err := strconv.ParseInt(userIDString, 10, 64)
	if err != nil {
		p.failOnErr(errors.New("用户登录状态出错"))
	}

	message := datamodels.NewMessage(productID, userID)
	byteMessage, err := json.Marshal(message)
	if err != nil {
		p.failOnErr(err)
	}

	// simple 模式
	p.RabbitMQ.Publish(string(byteMessage))
	return []byte("true")

	//product, err := p.ProductService.GetProductByID(productID)
	//if err != nil {
	//	return p.failOnErr(err)
	//}
	//
	//var orderID int64
	//showMessage := "抢购失败"
	//// 判断商品数量
	//if product.ProductNum > 0 {
	//	// 扣除商品数量
	//	product.ProductNum -= 1
	//	err := p.ProductService.UpdateProduct(product)
	//	if err != nil {
	//		return p.failOnErr(err)
	//	}
	//
	//	// 创建订单
	//	order := &datamodels.Order{
	//		UserID:      userID,
	//		ProductID:   productID,
	//		OrderStatus: datamodels.OrderSuccess,
	//	}
	//	orderID, err = p.OrderService.InsertOrder(order)
	//	if err != nil {
	//		return p.failOnErr(err)
	//	}
	//
	//	showMessage = "抢购成功"
	//}

	//return mvc.View{
	//	Layout: "shared/productLayout.html",
	//	Name:   "product/result.html",
	//	Data: iris.Map{
	//		"orderID":     orderID,
	//		"showMessage": showMessage,
	//	},
	//}
}
