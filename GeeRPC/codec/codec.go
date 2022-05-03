package codec

import "io"

/**
 * 消息头
 */
type Header struct {
	ServiceMethod string // format "Service.Method"
	Seq           uint64 // 客户端请求的序号（ID），用于区分不同的请求
	Error         string // 服务端返回的 Error
}

/**
 * 对消息体进行编解码的接口（Header + Body）
 */
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error // body 必须是指针
	Write(*Header, interface{}) error
}

type Type string

const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json"
)

type NewCodecFunc func(io.ReadWriteCloser) Codec

/**
 * 保存不同序列化协议对应的编解码接口的构造函数（而非实例）
 */
var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
