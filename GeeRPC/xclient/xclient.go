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
	mu      sync.Mutex // 用于 clients 的互斥
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
			mu.Unlock()
		}(rpcAddr)
	}

	wg.Wait()
	return e
}
