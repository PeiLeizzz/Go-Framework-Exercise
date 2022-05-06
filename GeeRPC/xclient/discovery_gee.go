package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

/**
 * 依赖于服务中心的对应的发现模块实例
 */
type GeeRegistryDiscovery struct {
	// 嵌套了不依赖服务中心的实例，复用其 Get 和 GetAll
	*MultiServersDiscovery
	// 注册中心的地址
	registry string
	// 服务列表的过期地址
	timeout time.Duration
	// 最后从服务中心更新服务列表的时间，默认 10s 过期，过期之后需要从注册中心拉取新的可用服务列表
	lastUpdate time.Time
}

const defaultUpdateTimeout = time.Second * 10

func NewGeeRegistryDiscovery(registry string, timeout time.Duration) *GeeRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	d := &GeeRegistryDiscovery{
		MultiServersDiscovery: NewMultiServersDiscovery(make([]string, 0)),
		registry:              registry,
		timeout:               timeout,
	}
	return d
}

func (d *GeeRegistryDiscovery) Refresh() error {
	// TODO：这里的锁是不是加在下面比较好
	d.mu.Lock()
	defer d.mu.Unlock()
	// 如果服务列表还没过期则不拉取
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	log.Println("rpc discovery: refresh servers from registry", d.registry)
	resp, err := http.Get(d.registry)
	if err != nil {
		log.Println("rpc discovery: refresh from registry err:", err)
		return err
	}
	servers := strings.Split(resp.Header.Get("X-Geerpc-Servers"), ",")
	d.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		// 删除前后的 " "
		if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, strings.TrimSpace(server))
		}
	}
	d.lastUpdate = time.Now()
	return nil
}

func (d *GeeRegistryDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

func (d *GeeRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	return d.MultiServersDiscovery.Get(mode)
}

func (d *GeeRegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultiServersDiscovery.GetAll()
}
