package common

import "net/http"

type FilterHandle func(w http.ResponseWriter, req *http.Request) error

type Filter struct {
	// 用来存储需要拦截的 URI
	filterMap map[string]FilterHandle
}

func NewFilter() *Filter {
	return &Filter{
		filterMap: make(map[string]FilterHandle),
	}
}

// 注册拦截器
func (f *Filter) RegisterFilterUri(uri string, handler FilterHandle) {
	f.filterMap[uri] = handler
}

func (f *Filter) GetFilterHandle(uri string) FilterHandle {
	return f.filterMap[uri]
}

type WebHandle func(w http.ResponseWriter, req *http.Request)

func (f *Filter) Handle(webHandle WebHandle) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if handle, ok := f.filterMap[req.RequestURI]; ok {
			err := handle(w, req)
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}
		}

		// 执行正常注册的函数
		webHandle(w, req)
	}
}
