package geerpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

const MagicNumber = 0x3bef5c

/**
 * 用于协商消息的编解码方式
 * 对于 Option 自身的内容，默认采用固定的 JSON 来编解码
 * | Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
 * | <------      固定 JSON 编码      ------>  | <--------   编码方式由 CodeType 决定   -------->|
 * 为了提升性能，一般在报文的最开始会规划固定的字节，来协商相关的信息。
 * 比如第1个字节用来表示序列化方式，第2个字节表示压缩方式，第3-6字节表示 header 的长度，7-10 字节表示 body 的长度。
 * 在这里仅协商消息的编解码方式。
 * 在一次连接中，Option 固定在报文的最开始，Header 和 Body 可以有多对，例如：
 * | Option | Header1 | Body1 | Header2 | Body2 | ...
 */
type Option struct {
	// 标识 geerpc
	MagicNumber int
	// 标识选择的编解码方式
	CodecType codec.Type
	// 连接的超时时间，如果为 0 代表不限制
	ConnectTimeout time.Duration
	// 处理的超时时间
	HandleTimeout time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: time.Second * 10,
}

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go server.ServeConn(conn)
	}
}

/**
 * 可以快捷地启动服务
 * lis, _ := net.Listen("tcp", ":9999")
 * geerpc.Accept(lis)
 */
func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}

/**
 * 通过传入的 Type.Method，解析出 service 的名称和 method 的名称
 * 然后从 server 的 serviceMap 中取出对应的 service 实例
 * 再从该实例中取出 method 实例
 */
func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svci.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}

/**
 * 先处理 Option，再处理 request
 */
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()

	// time.Sleep(time.Second * 5)
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options decode error:", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}

	server.serveCodec(f(conn), &opt)
}

// a placeholder for response argv when error
var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc codec.Codec, opt *Option) {
	sending := new(sync.Mutex) // 保证发送完整的响应
	wg := new(sync.WaitGroup)  // 等待所有请求都被处理
	// 一个连接可能有多个并行的请求（opt 后接多对 header body）
	for {
		req, err := server.readRequest(cc) // 读取请求
		if err != nil {
			if req == nil {
				break // 无请求出错时不用 recover + response，直接退出循环即可（例如连接被关闭了）
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending) // 回复请求
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg, opt.HandleTimeout) // 处理请求
	}
	wg.Wait()
	_ = cc.Close()
}

type request struct {
	// request header
	h *codec.Header
	// request body
	argv reflect.Value
	// reply body
	replyv reflect.Value
	mtype  *methodType
	svc    *service
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request{h: h}
	req.svc, req.mtype, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReplyv()

	// 确保 argvi 是可取地址的，ReadBody 需要指针类型（反序列化存储 client 传来的入参）
	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read argv err:", err)
		return req, err
	}
	return req, nil
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	// 这里要加锁的原因是，一个 cc.Conn 可能导致多个 sendResponse 协程在同时工作
	// 并且这些个 Conn 都是同一条连接，不能发生写冲突
	// 即防止协程交错响应客户端
	// 但是 Go 的文件描述符（FD）的写入已保证线程安全
	// 这里另一个原因是 cc.buf 的存在
	// 对于 cc.buf.Flush() 时，避免其他协程再向该缓冲区中写数据造成冲突
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) handleRequest(cc codec.Codec, req *request,
	sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		err := req.svc.call(req.mtype, req.argv, req.replyv) // call: svc.mtype(argv, replyv)
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}
		server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
		sent <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}
	select {
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("rpc server: request handle timeout: expect within %s", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
		// TODO: 这里直接退出，那上面协程会不会被阻塞？(goroutine 泄漏)
	case <-called:
		<-sent
	}
}

// ----- 支持 HTTP -----

const (
	connected        = "200 Connected to Gee RPC"
	defaultRPCPath   = "/_geerpc_"
	defaultDebugPath = "/debug/geerpc"
)

/**
 * 用于处理 client 发来的 CONNECT 请求，并且将后续的通信转换为 RPC
 * 如连接成功，则返回正确响应，并且通过 HTTP 下层的 TCP 与 client 建立 RPC 连接（切换为 RPC）
 * 该函数本身在 net/http 包的处理中就是在协程中的，
 * 因此 RPC 连接的效果与 Accept() 中一样，可以支持多个客户端
 */
func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 must CONNECT\n")
		return
	}
	// 获取 TCP 套接字，接管 HTTP 连接
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking", req.RemoteAddr, ":", err.Error())
		return
	}
	// 返回连接成功的响应
	_, _ = io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	// 切换为 RPC 形式与 client 建立连接（与 Accept() 中的 go server.ServeConn 效果一样）
	server.ServeConn(conn)
}

/**
 * HTTP 路由处理器（与 Accept() 是相互独立的）
 * defaultRPCPath -> server 处理与远程地址的连接（客户端可以向该地址发送 CONNECT 请求），之后与客户端建立 RPC 连接
 * 其他路径可以提供一些别的功能（例如日志、统计等）
 * e.g. defaultDebugPath -> debugHTTP 处理逻辑
 */
func (server *Server) HandleHTTP() {
	http.Handle(defaultRPCPath, server)
	http.Handle(defaultDebugPath, debugHTTP{server})
	log.Println("rpc server debug path:", defaultDebugPath)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
