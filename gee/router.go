package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node       // 存储每种请求方式的 Trie 树根节点
	handlers map[string]HandlerFunc // 存储每个 pattern 的方法
}

// roots key: ['GET'], ['POST']
// handlers key: ['GET-/p/:lang/doc'], ['POST-/p/book']
func newRouter() *router {
	return &router{
		roots:    make(map[string]*node),
		handlers: make(map[string]HandlerFunc),
	}
}

func parsePattern(pattern string) []string {
	// '/p/:lang/doc' -> ['p', ':lang', 'doc']
	vs := strings.Split(pattern, "/") // 不能直接用，要先预处理
	// 比如 "/v1/" -> ["", "v1", ""] 要把空字符串的元素去掉

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			if item[0] == '*' { // 遇到第一个 *，后面就不看了
				break
			}
		}
	}
	return parts
}

func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)

	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	searchParts := parsePattern(path)
	params := make(map[string]string)
	root, ok := r.roots[method]

	if !ok {
		return nil, nil
	}

	n := root.search(searchParts, 0)

	if n != nil {
		parts := parsePattern(n.pattern)
		for index, part := range parts {
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
			} else if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

/**
 * 解析请求的路径，查找路由映射表
 * 如果找到，就执行注册的处理方法
 * 如果找不到，返回 404 NOT FOUND
 */
func (r *router) handle(c *Context) {
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		// 将路由匹配的 Handler 添加到中间件列表中
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
		})
	}
	// 依次执行中间件
	c.Next()
}
