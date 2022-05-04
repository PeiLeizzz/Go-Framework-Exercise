package geerpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"sync"
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
	return newClientCodec(f(conn), opt), nil
}

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

func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}

	// 如果 NewClient 中 client 建立出错（nil）
	// 就把 conn 关闭
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, opt)
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
 */
func (client *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <-client.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}
