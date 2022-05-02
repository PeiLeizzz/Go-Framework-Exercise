## 一个 Gin 的 Demo

- 用到的技术有：`gin`、`gorm`、`jwt-go`、`ini`、`swagger`、`redis`、`docker`、`nginx`

### 2022-4-27 

#### 热更新

- 热更新目的
  1. 不关闭现有连接（正在运行中的程序）
  2. 新的进程启动并代替旧进程
  3. 新的进程接管**新**的连接
  4. 旧用户仍在请求旧连接时要保持连接，新用户应请求新进程，不可拒绝请求
- 热更新重启流程：
  1. 替换可执行文件或修改配置
  2. 发送信号量 `SIGHUP`，进程可以捕捉该信号并执行相关的处理函数
  3. 拒绝新连接请求旧进程，但要保证旧连接正常
  4. 启动新的子进程
  5. 新的子进程开始 `Accet`
  6. 系统将新的连接请求转交给新的子进程
  7. 旧进程处理完所有旧连接后正常结束
- **优雅的重启**：使用 `endless` 实现零停机重新启动服务，`endless server` 监听以下几种信号量
  1. `syscall.SIGHUP`：触发 `fork` 子进程和重新启动（`kill -1 pid`）
  2. `syscall.SIGUSR1 / syscall.SIGTSTP`：被监听，但不会触发任何动作
  3. `syscall.SIGUSR2`：触发 `hammerTime`
  4. `syscall.SIGINT / syscall.SIGTERM`：触发服务器关闭（会完成正在运行的请求）

  > `endless` 热更新：创建子进程后将原进程退出，有点不符合守护进程的要求
  >
- **优雅的关闭**：`http.Server` 的 `Shutdown` 方法

### 2022-4-28

#### Docker

- Docker 是一个开源的轻量级容器技术，让开发者可以打包他们的应用以及应用运行的上下文环境到一个可移植的镜像中，然后发布到任何支持 Docker 的系统上运行。通过容器技术，在几乎没有性能开销的情况下，Docker 为应用提供了一个隔离运行环境。
- 简化配置
- 代码流水线管理
- 提高开发效率
- 隔离应用
- 快速、持续部署
- MySQL 创建用户 + 密码
  1. `CREATE USER 'root'@'%' IDENTIFIED BY 'root';`
  2. `GRANT ALL ON *.* TO 'root'@'%';`
  3. `ALTER USER 'root'@'%' IDENTIFIED BY 'password';`
- Docker 命令
  1. 启动停止的容器 `docker start id`
  2. 查看容器 `docker ps -a`
  3. 创建 mysql 容器 `docker run --name mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=password -v absolutepath:/var/lib/mysql -d mysql/mysql-server`
  4. 创建关联了 mysql 的应用容器 `docker run --link mysql:mysql -p 8000:8000 gin-blog-docker-scratch`

#### 交叉编译

- Golang 跨平台编译：`CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-gin-example .`
  - `CGO_ENABLED=0`：用于声明 cgo 工具不可用（存在交叉编译的情况时，cgo 不可用），关闭 cgo 后，在构建过程中会忽略 cgo 并**静态链接**所有的依赖库，而开启 cgo 后，方式将转为动态链接
  - `GOOS=linux`：用于标识（声明）程序构建环境的目标操作系统
  - `GOARCH`：用于标识（声明）程序构建环境的目标计算架构
  - `GOHOSTOS`：用于标识（声明）程序运行环境的目标操作系统
  - `GOHOSTARCH`：用于标识（声明）程序运行环境的目标计算架构
  - `go build`
    - `-a`：强制重新编译
    - `-installsuffix`：在软件包安装的目录中增加后缀标识，以保持输出与默认版本分开
    - `-o`：指定编译后的可执行文件名称

### 2022-4-30

#### Redis

- Redis 命令
  1. 启动 redis 服务 `brew services start redis` / `redis-server /usr/local/etc/redis.conf`
  2. 关闭 redis 服务 `redis-cli shutdown`
- TODO: Redis 和 MySQL 的双写一致性

### 2022-5-2

#### Make

- Make 是一个构建自动化工具，会在当前目录下寻找 Makefile 或 makefile 文件，依据该文件的构建规则去完成构建
- Makefile 规则：

  - Makefile 由多条规则组成，每条规则都以一个 target（目标）开头，后跟一个 `:` 冒号，冒号后是这一个目标的 prerequisites（前置条件）
  - 紧接着新的一行，**必须以一个 tab 作为开头**，后面跟随 command（命令），也就是你希望这一个 target 所执行的构建命令

  ```makefile
  [target] ...: [prerequisites] ...
  <tab>[command]
      ...
      ...
  ```

  - target：一个目标代表一个规则，可以是一个或多个文件名，也可以是某个操作的名字（标签），称为**伪目标**（phony）
  - prerequisites：前置条件，这一项是**可选**参数，通常是多个文件名、伪目标。它的作用是 target 是否需要重新构建的标准，如果**前置条件不存在或有过更新（文件的最后一次修改时间）则认为 target 需要重新构建**
  - command：构建这一个 target 的具体命令集
  - `.PHONY`：声明后面的标签为伪目标
    - 声明为伪目标后：在执行对应的命令时，make 就不会去检查标签名对应的文件，而是每次都会运行标签对应的命令
    - 若不声明：恰好存在对应的文件，则 make 将会认为 xx 文件已存在，没有重新构建的必要了
- 在编写 Makefile 前，需要先分析构建先后顺序、依赖项，需要解决的问题等
- make 默认会打印每条命令，再执行。这个行为被定义为**回声**，可以在对应的命令前加上 `@`，指定该命令不被打印到标准输出
  ```makefile
  build:
    @go build -v .
  ```

#### Nginx
- Nginx 命令：
  1. `nginx`：启动
  2. `nginx -s stop`：立刻停止
  3. `nginx -s reload`：重新加载配置文件
  4. `nginx -s quit`：平滑停止
  5. `nginx -t`：测试配置文件是否正确，同时也可以用来查看配置文件所在位置
  6. `nginx -v`：显示 Nginx 版本信息
  7. `nginx -V`：显示 Nginx 版本信息、编译器和配置参数的信息
- 涉及配置：
  1. `proxy_pass`：配置反向代理的路径（如果 `proxy_pass` 的 `url` 最后为 `/` 表示绝对路径，会去掉匹配的前缀，否则为相对路径）
       - ```nginx
          location /proxy {
            proxy_pass http://192.168.137.181:8080/
          }
          ```
          当访问 `http://127.0.0.1/proxy/test/test.txt` 时，会被转发到 `http://192.168.137.181:8081/test/test.txt`
       - ```nginx
          location /proxy {
            proxy_pass http://192.168.137.181:8080
          }
          ```
          当访问 `http://127.0.0.1/proxy/test/test.txt` 时，会被转发到 `http://192.168.137.181:8081/proxy/test/test.txt`
  2. `upstream`：配置**负载均衡**，默认以**轮询**的方式进行负载，还支持四种模式：
     1. `weight`：权重，指定轮询的概率，`weight` 与访问概率成正比
     2. `ip_hash`：按照访问 IP 的 hash 结果值分配
     3. `fair`：按后端服务器响应时间进行分配，响应时间越短优先级别越高
     4. `url_hash`：按照 URL 的 hash 结果值分配
    具体配置：例如启动了两个端口的服务 `127.0.0.1:8001` 和 `127.0.0.1:8002`，在 `nginx.conf` 的 `upstream` 节点中进行如下配置：
      ```nginx
      http {
        # ...
        upstream api.blog.com {
          server 127.0.0.1:8081;
          server 127.0.0.1:8082;
        }

        server {
          listen: xxxx;
          server_name: xxxx;

          location / {
            # http:// + upstream 的节点名称
            proxy_pass http://api.blog.com/;
          }
        }
      }
      ```