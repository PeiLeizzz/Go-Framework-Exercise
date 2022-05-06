package xclient

import (
	"context"
	. "geerpc"
	"io"
	"reflect"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *Option
	mu      sync.RWMutex // 用于 clients 的互斥
	clients map[string]*Client
}

var _ io.Closer = (*XClient)(nil)

func NewXClient(d Discovery, mode SelectMode, opt *Option) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opt:     opt,
		clients: make(map[string]*Client),
	}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, client := range xc.clients {
		// ignore error
		_ = client.Close()
		delete(xc.clients, key)
	}
	return nil
}

func (xc *XClient) dial(rpcAddr string) (*Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	// TODO: 这里的加锁导致了串行化的连接
	// 是否可以优化（因为可能 XDial 很耗时、甚至无限等待，这样就直接造成后面的 client 无法去连接 server）
	// 想法：加一个 xc.map[cid]chan *client，每个新的 client 连接请求对应一个 cid
	// 如果一个 client == nil，同时它的 cid（可以根据地址唯一化） 存在于 map 中（说明可能另一个协程中它正在与 server 尝试连接，但还未完成）
	// 此时让这个 client 等待（client <-[cid])，当另一边连接完成后（[cid]<- client）
	// 如果一个 client == nil，但它的 cid 不存在于 map 中，那么就将这个 cid 加入 map，让它去连接 server，连接完成后，[cid]<- client，同时在 map 中删去这个 cid，并将连接存入 xc.clients
	// 这样的好处在于，不同的 client 可以同时去尝试连接 server，而不用被上一个 client 的连接请求阻塞
	// 这样的问题在于，如果好多相同的 client 同时在接收这个 channel，缓冲区大小又是不确定的，只有一个发送方，如何防止阻塞？
	client, ok := xc.clients[rpcAddr]
	if ok && !client.IsAvailable() {
		_ = client.Close()
		delete(xc.clients, rpcAddr)
		client = nil
	}
	// 如果找不到该地址对应的 client，或者原先缓存的 client 连接已不可用，则与服务端创建新的连接，缓存并返回新 client
	if client == nil {
		var err error
		client, err = XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcAddr] = client
	}
	return client, nil
}

/**
 * 与 Client.Call() 不同的是，这个函数可以接收服务端地址，包含建立连接的 Dial()
 */
func (xc *XClient) call(rpcAddr string, ctx context.Context,
	serviceMethod string, args, reply interface{}) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

/**
 * 根据负载均衡策略，从发现模块自动选择服务端实例（如果未有客户端连接则先创建客户端、连接服务端）、发送请求
 */
func (xc *XClient) Call(ctx context.Context, serviceMethod string,
	args, reply interface{}) error {
	// 根据负载均衡策略，从发现模块自动选择服务端实例
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}

/**
 * 将 Client 请求的函数发送给所有注册的服务端处理
 */
func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string,
	args, reply interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var e error
	// 如果不需要响应，直接为 true
	replyDone := reply == nil
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply interface{}
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(rpcAddr, ctx, serviceMethod, args, clonedReply)
			// 需要使用互斥锁保证 e 和 reply 被正确修改
			mu.Lock()
			defer mu.Unlock()
			// 如果任意一个实例发生错误，则返回其中一个错误（最早发生的）
			if err != nil && e == nil {
				e = err
				cancel() // 有错误发生时，快速失败，作用于 Client.Call() 中
			}
			// 如果调用成功，则返回其中一个的结果（最早成功的）
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
		}(rpcAddr)
	}

	wg.Wait()
	return e
}
