### gRPC
- gRPC 基于 HTTP/2 标准设计，拥有双向流、流控、头部压缩、单 TCP 连接上的多复用请求等特性，在移动设备上表现更好、节省空间
- gRPC 优点：
  1. 性能好：Protobuf 数据描述文件小，解析速度快
  2. 代码生成方便：可从 proto 文件中生成服务基类、消息体、客户端等代码，客户端与服务端共用一个 proto 文件
  3. 流传输
  4. 超时和取消
- gRPC 缺点：
  1. 可读性差：Protobuf 序列化后是二进制格式的数据
  2. 不支持浏览器调用
  3. 外部组件支持较差
- 四种调用方式
  1. Unary RPC：一元 RPC，就是客户端发起的一次普通 RPC 请求
  2. Server-side streaming RPC：服务端流式 RPC，这是一个单向流，客户端发起一次 RPC 请求，服务端通过流式响应多次发送数据集，客户端 Recv 接收数据集
  3. Client-side streaming RPC：客户端流式 RPC，这是一个单向流，客户端通过流式发起多次 RPC 请求给服务端，而服务端仅发起一次响应给客户端
  4. Bidirectional streaming RPC：双向流式 RPC，这是一个双向流，客户端以流式的方式发起请求，服务端同样以流式的方式响应请求。**首个请求一定是由客户端发起的**，但具体的交互方式（谁先谁后、一次发多少、响应多少、什么时候关闭）是由程序编写的方式来确定（可以结合协程）
- Unary RPC 缺点：
  1. 数据包过大会造成瞬时压力
  2. 不能做到客户端边发送，服务端边处理（必须所有数据包都接收成功且正确后，才能回调响应，进而进行业务处理）
- Streaming RPC 优点：
  1. 适用于大数据包场景
  2. 可以实时交互
- 相关通信协议：
  1. gRPC 在建立连接之前，客户端和服务端会发送连接前言（Magic + SETTING），以确定协议和配置项
  2. gRPC 在传输数据时，会涉及到滑动窗口（WINDOW_UPDATE）等流控策略
  3. 在传输 gRPC 附加信息时，是基于 HEADERS 帧进行传播和设置的，具体的请求/响应数据存储在 DATA 帧中
  4. gRPC 的请求/响应结果可分为 HTTP 和 gRPC 状态响应（grpc-status、grpc-message）两种类型
  5. PING/PONG 用于判断当前连接是否可用，常用于计算往返时间
  
### Protobuf
- Protobuf（Protocol Buffers）是一种与语言、平台无关，且可扩展的序列化结构化数据的数据描述语言，通常称其为 IDL，常用于通信协议、数据存储等，与 JSON、XML 相比，它更小、更快。
  ```protobuf
  // 声明使用 proto3 语法（不声明默认使用 proto2）
  syntax = "proto3";

  package helloworld;

  // 定义名为 Greeter 的 RPC 服务（service）
  service Greeter {
      // rpc 方法名（入参）returns（出参）
      rpc SayHello (HelloRequest) returns (HelloReply) {}
  }
  
  // 定义消息体
  message HelloRequest {
      // 类型、字段名称、字段编号
      string name = 1;
  }

  message HelloReply {
      string message = 1;
  }
  ```
- 项目根目录下通过 grpc 生成 go 代码：
  ```bash
  $ protoc --go_out=plugins=grpc:. ./proto/*.proto
  ```