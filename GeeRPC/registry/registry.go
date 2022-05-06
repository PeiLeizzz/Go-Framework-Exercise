package registry

import (
	"errors"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// TODO：缺少服务地址与相关服务的映射
type GeeRegistry struct {
	timeout time.Duration // 超时时间，如果为 0 代表不限制
	mu      sync.Mutex
	servers map[string]*ServerItem // 已注册的所有服务实例，键为服务端地址
}

type ServerItem struct {
	Addr  string
	start time.Time
}

func New(timeout time.Duration) *GeeRegistry {
	return &GeeRegistry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

const (
	defaultPath    = "/_geerpc_/registry"
	defaultTimeout = time.Minute * 5
)

var DefaultGeeRegister = New(defaultTimeout)

/**
 * 添加服务实例 / 如果服务已经存在，则更新其 start
 */
func (r *GeeRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		s.start = time.Now()
	}
}

/**
 * 删除服务实例
 */
func (r *GeeRegistry) removeServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.servers[addr]
	if s != nil {
		delete(r.servers, addr)
	}
}

/**
 * 检查所有服务，删除超时的服务，并返回所有可用的服务地址
 */
func (r *GeeRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		// 如果不限制超时时间或者还未超时
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			// 如果超时了，则删除
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

/**
 * 注册中心采用 HTTP 协议提供服务，所有有用的信息都承载在 HTTP Header 中
 * 支持两种 HTTP Method，都通过 Header 自定义字段 X-Geerpc-Servers 承载：
 * 1. GET：返回所有可用的服务列表
 * 2. POST：添加服务实例或发送心跳
 * 3. DELETE: 删除服务实例（一般是服务发送心跳出错时）
 */
func (r *GeeRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("X-Geerpc-Servers", strings.Join(r.aliveServers(), ","))
	case "POST":
		addr := req.Header.Get("X-Geerpc-Server")
		// 服务器发来的自己的地址
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	case "DELETE":
		addr := req.Header.Get("X-Geerpc-Server")
		// 服务器发来的自己的地址
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.removeServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *GeeRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path:", registryPath)
}

// 默认的注册中心实例，监听 /_geerpc_/registry 路径
func HandleHTTP() {
	DefaultGeeRegister.HandleHTTP(defaultPath)
}

/**
 * 用于服务端向注册中心注册自己，并定时向注册中心发送心跳
 */
func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		// 默认发送心跳的周期比注册中心设置的默认过期时间少 1 min
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
			err = errors.New("test")
		}
		t.Stop()
		removeServer(registry, addr)
	}()
}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-Geerpc-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: heart beat err:", err)
		return err
	}
	return nil
}

func removeServer(registry, addr string) error {
	log.Println(addr, "send heart goroutine break, remove itself from the registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("DELETE", registry, nil)
	req.Header.Set("X-Geerpc-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: remove itself err:", err)
		return err
	}
	return nil
}
