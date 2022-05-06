package geerpc

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

/**
 * 一个函数能够被远程调用的条件：
 * 1. 方法的类型(接收器类型)是导出的
 * 2. 方法是导出的
 * 3. 方法有两个参数，它们的类型都是导出的
 * 4. 方法的第二个参数是指针（承载返回值）
 * 5. 方法返回值是 error
 *
 * e.g. func (t *T) MethodName(arg T1, reply *T2) error
 */

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{} // arguments to the function
	Reply         interface{} // reply from the function(pointer)
	Error         error       // if error occurs, it will be set
	Done          chan *Call  // Strobes when call is complete
}

// 调用结束时，通知调用方
func (call *Call) done() {
	call.Done <- call
}

/**
 * 一个 Client 可能被多个协程同时使用
 */
type Client struct {
	cc       codec.Codec // 消息的编解码器
	opt      *Option
	sending  sync.Mutex       // 保证数据有序发送的互斥锁
	header   codec.Header     // 每个请求的消息头（请求发送是互斥的，因此只需要一个头）
	mu       sync.Mutex       // 保证 Client 自身状态的互斥更改
	seq      uint64           // 请求编号
	pending  map[uint64]*Call // 存储未处理完的请求，键是编号，值是请求 Call 实例
	closing  bool             // user has called Close（一般是用户主动关闭的）
	shutdown bool             // server has told us to stop（一般是有错误发生）
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("connection is shut down")

func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

// 返回 client 是否正在工作
func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closing
}

func (client *Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	// 不用 IsAvaiable，不可重入锁，如果分成两次上锁会影响性能
	if client.closing || client.shutdown {
		return 0, ErrShutdown
	}
	defer client.mu.Unlock()
	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

// 服务端或客户端发生错误时调用，并且将错误信息通知所有 pending 状态的 call
func (client *Client) terminateCalls(err error) {
	// TODO: 为什么这里要加 sending 锁？
	// 防止产生错误要终止 client 时，还有新的消息在 send()
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()

	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}

/**
 * 接收服务端响应
 * 三种情况：
 * 1. call 不存在，可能是请求没有发送完整，或者因为其他原因取消，但是服务端仍然处理并响应了
 * 2. call 存在，但服务端处理出错，即返回的 h.Error 不为空
 * 3. call 存在，服务端处理正常，需要从返回的 body 中读取 Reply
 */
func (client *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = client.cc.ReadHeader(&h); err != nil {
			break
		}
		call := client.removeCall(h.Seq)
		switch {
		case call == nil:
			err = client.cc.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = client.cc.ReadBody(nil)
			call.done()
		default:
			err = client.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}

	// err occurs
	client.terminateCalls(err)
}

/**
 * 客户端构造函数
 */
type newClientFunc func(conn net.Conn, opt *Option) (*Client, error)

/**
 * 向服务端发送 options，如果发送成功，创建 client 并开启 receive 协程
 */
func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}

	// 发送 options
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error:", err)
		_ = conn.Close()
		return nil, err
	}
	// 接收 options 响应，防止 options 粘包
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc client: read options response error:", err)
		_ = conn.Close()
		return nil, err
	}
	log.Println("rpc client: exchange options success, you can send some request now")
	return newClientCodec(f(conn), opt), nil
}

/**
 * 建立 HTTP 连接，之后转换为 RPC 协议
 * 向服务端发送 CONNECT 报文，如果成功收到响应，则创建客户端、启动 receive 协程
 * 连接建立的逻辑：
 * NewHTTPClient -> 发送 CONNECT HTTP 报文 -> 接收响应，并且响应的 Status 正确 ->
 * （切换为 RPC 协议）NewClient -> 发送 options -> newClientCodec -> receive
 */
func NewHTTPClient(conn net.Conn, opt *Option) (*Client, error) {
	// 发送 CONNECT 报文
	_, _ = io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", defaultRPCPath))

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	defer func() {
		_ = resp.Body.Close()
	}()

	// 通过 HTTP CONNECT 请求建立连接后，后续通信过程切换为 RPC 协议
	if err == nil && resp.Status == connected {
		log.Println("rpc client: connected with http rpc server")
		return NewClient(conn, opt)
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	return nil, err
}

/**
 * 创建客户端并且启动 receive 协程
 */
func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1, // starts with 1, 0 means invalid call
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

func parseOptions(opts ...*Option) (*Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

/**
 * 储存客户端创建的结果
 */
type clientResult struct {
	client *Client
	err    error
}

/**
 * 连接服务端，并且处理连接超时
 */
func dialTimeout(f newClientFunc, network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}

	// 建立 TCP 连接
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}

	// 如果 NewClient 中 client 建立出错（nil）
	// 就把 conn 关闭
	defer func() {
		if err != nil {
			_ = conn.Close()
		}
	}()

	ch := make(chan clientResult, 1)
	go func() {
		client, err := f(conn, opt)
		ch <- clientResult{
			client: client,
			err:    err,
		}
	}()

	if opt.ConnectTimeout == 0 {
		result := <-ch
		return result.client, result.err
	}

	select {
	case <-time.After(opt.ConnectTimeout):
		// TODO: 超时退出后上面协程中的 ch 可能被阻塞，会导致泄漏
		// 修改：加了缓冲区，是否会有问题？
		return nil, fmt.Errorf("rpc client: connect timeout: expect within %s", opt.ConnectTimeout)
	case result := <-ch:
		return result.client, result.err
	}
}

func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	return dialTimeout(NewClient, network, address, opts...)
}

func DialHTTP(network, address string, opts ...*Option) (*Client, error) {
	return dialTimeout(NewHTTPClient, network, address, opts...)
}

/**
 * 客户端连接服务端的统一入口，支持不同协议
 * rpc server 地址格式为：protocol@addr
 * e.g. http@10.0.0.1:7001, tcp@10.0.0.1:9999, unix@/tmp/geerpc.sock
 */
func XDial(rpcAddr string, opts ...*Option) (*Client, error) {
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("rpc client err: wrong format '%s', expect protocol@addr", rpcAddr)
	}
	protocol, addr := parts[0], parts[1]
	switch protocol {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		// tcp, unix or other transport protocol
		return Dial(protocol, addr, opts...)
	}
}

/**
 * 客户端发送请求
 */
func (client *Client) send(call *Call) {
	client.sending.Lock()
	defer client.sending.Unlock()

	// 注册
	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// 准备请求头
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	// 编码、发送
	if err := client.cc.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		// 即使产生了 err，call 也可能是 nil
		// 这表示，部分写入失败了，但是服务器还是收到了请求
		// 并且返回了响应，而客户端已经在 receive 中将该响应处理了
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

/**
 * 异步接口，直接返回 call（即使 call 还未完成）
 * done 缓冲区的作用：可以给多个 client.Go 传入同一个 done 对象，
 * 从而控制异步请求并发的数量（可以阻塞 receive 中的 call.done()）
 * 但不会阻塞发送（send）
 *
 * e.g. 异步使用
 *     call := client.Go( ... )
 *     go func(call *Call) {
 *         select {
 *             <- call.Done:
 *                 // do something
 *             <- otherChan:
 *                 // do something
 *         }
 *     }(call)
 *
 *     otherFunc()
 */
func (client *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	go client.send(call)
	return call
}

/**
 * 同步接口，会被阻塞
 * Call 会等待函数执行完成 Call -> Go -> send -> call.done
 * 超时处理通过 context 来完成，交给用户控制
 * e.g.
 *     ctx, _ := context.WithTimeout(context.Background(), time.Second)
 *     var reply int
 *     err := client.Call(ctx, "Foo.Sum", &Args{1, 2}, &reply)
 *     // ...
 */
func (client *Client) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	call := client.Go(serviceMethod, args, reply, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		client.removeCall(call.Seq)
		// 若是该函数退出后 receive 又收到了响应，call.done() 岂不是会阻塞导致 receive 被阻塞？
		// 其实不会，因为 call.Done 有至少 一个单位 的缓冲区，call.done() 可以正常执行
		// 那下一次的 Call 岂不是会秒返回？不会，Call() 函数中每个 call 对应的缓冲区都是新建的，不存在复用
		return errors.New("rpc client: call failed: " + ctx.Err().Error())
	case call := <-call.Done:
		return call.Error
	}
}
