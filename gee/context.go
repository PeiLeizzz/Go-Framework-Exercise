package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// 方便构建 JSON 数据
type H map[string]interface{}

/**
 * Context 随着每一个请求的出现而产生，请求的结束而销毁，
 * 和当前请求强相关的信息都应由 Context 承载。
 * 因此，设计 Context 结构，扩展性和复杂性留在了内部，
 * 而对外简化了接口。路由的处理函数，以及将要实现的中间件，
 * 参数都统一使用 Context 实例， Context 就像一次会话的百宝箱，
 * 可以找到任何东西。
 */
type Context struct {
	Writer http.ResponseWriter
	Req    *http.Request

	// request 信息
	Path   string
	Method string
	Params map[string]string // 路由参数访问：例如 /p/:lang/doc -> /p/go/doc
	// {"lang": "go"}

	// response 信息
	StatusCode int

	// middleware
	handlers []HandlerFunc
	index    int // 记录当前执行到第几个中间件

	engine *Engine
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path, // 不包含参数信息
		// 例如地址 /?name=123
		// req.URL.Path = /
		// req.RequestURI = /?name=123 注意区分
		Method: req.Method,
		index:  -1,
	}
}

/**
 * 类似于洋葱模型
 * func A(c *Context) {
 *     part1
 * 	   c.Next()
 * 	   part2
 * }
 *
 * func B(c *Context) {
 *     part3
 * 	   c.Next()
 *     part4
 * }
 *
 * 执行顺序：part -> part3 -> handler -> part4 -> part2
 */
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

func (c *Context) Fail(code int, err string) {
	c.index = len(c.handlers) // 短路中间件的执行
	c.JSON(code, H{
		"message": err,
	})
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, name string, data interface{}) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	if err := c.engine.htmlTemplates.ExecuteTemplate(c.Writer, name, data); err != nil {
		c.Fail(500, err.Error())
	}
}
