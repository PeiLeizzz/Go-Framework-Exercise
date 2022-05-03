package geerpc

import (
	"encoding/json"
	"fmt"
	"geerpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
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
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Server struct{}

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

	server.serveCodec(f(conn))
}

// a placeholder for response argv when error
var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc codec.Codec) {
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
		go server.handleRequest(cc, req, sending, wg) // 处理请求
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
	// TODO: 需要获取 argv 的类型，这里先假定都为 string
	// argv 是可取地址的类型
	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
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

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	// TODO: call the registered rpc methods to get the right response
	defer wg.Done()
	log.Println(req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	// 这里把 request.header 直接作为 response.header 传回去了
	server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}
