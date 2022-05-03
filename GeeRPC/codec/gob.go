package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser // 通常是通过 TCP 或者 Socket 等得到的链接实例
	buf  *bufio.Writer      // 防止阻塞而创建的带缓冲的 Writer
	dec  *gob.Decoder
	enc  *gob.Encoder
}

var _ Codec = (*GobCodec)(nil)

// 构造函数
func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn) // 向 conn 中写，编码用
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn), // 从 conn 中读，解码用
		enc:  gob.NewEncoder(buf),  // 向 buf 中写 -> 向 conn 中写
	}
}

// ------------------------- 实现 Codec 接口部分 -------------------------

func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h) // 解码 conn，并写入 header
}

func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body) // 解码 conn，并写入 body（body 必须是指针）
}

func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	// 先写入 buf，再从 buf 写入 conn
	defer func() {
		_ = c.buf.Flush()
		if err != nil {
			_ = c.Close()
		}
	}()

	if err := c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}

	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}

	return nil
}

func (c *GobCodec) Close() error {
	return c.conn.Close()
}
