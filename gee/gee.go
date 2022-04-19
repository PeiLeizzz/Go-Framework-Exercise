package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

type HandlerFunc func(*Context)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc
	engine      *Engine
}

// 实现 ServeHTTP 接口的一个 Gee 实例
type Engine struct {
	*RouterGroup // 内嵌结构体，可以获取其方法
	// engine 实例的直接 group 的 prefix 为 ""
	router *router
	groups []*RouterGroup
	// html 模版渲染
	htmlTemplates *template.Template
	funcMap       template.FuncMap
}

// gee.Engine 的构造函数
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func Default() *Engine {
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var middlewares []HandlerFunc
	// 直接循环加入有些太暴力？应该跟前缀树结合？
	// 还有是不是应该去重？
	for _, group := range engine.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(
		template.New("").
			Funcs(engine.funcMap).
			ParseGlob(pattern))
}

// 添加中间件
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := path.Join(group.prefix, relativePath)
	// 去除 Request.URL.Path 的前缀，因为 fileServer 默认使用 Request.URL.Path 作为文件路径
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	return func(c *Context) {
		// 得到本地文件 root 下的相对路径
		file := c.Param("filepath")
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		// fileServer 是一个 HandlerFunc 类型
		// fileServer.ServeHTTP(c.Writer, c.Req)
		// -> fileServer(c.Writer, c.Req)
		// 内部实现：-> http.FileServer(fs).ServeHTTP(c.Writer, c.Req)
		fileServer.ServeHTTP(c.Writer, c.Req) // 需要手动调用
		// engine 的 ServeHTTP 不需要手动调用的原因是在于
		// 我们在 Run 函数的 ListenAndServe 中将其挂载了
	}
}

// 将静态文件夹本地根目录 root 映射到路由 relativePath/*filepath
// 将文件路径提供给静态文件服务器后，剩下的都叫给静态文件服务器的 ServeHTTP 来处理
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	group.GET(urlPattern, handler)
}
