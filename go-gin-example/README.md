## 一个 Gin 的 Demo

- 用到的技术栈有：`gin`、`gorm`、`jwt-go`、`ini`

### 2022-4-27 热更新

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
- **优雅的关闭**：`http.Server` 的 `Shutdown` 方法
  
### 2022-4-28 
#### Docker
-  Docker 是一个开源的轻量级容器技术，让开发者可以打包他们的应用以及应用运行的上下文环境到一个可移植的镜像中，然后发布到任何支持 Docker 的系统上运行。通过容器技术，在几乎没有性能开销的情况下，Docker 为应用提供了一个隔离运行环境。
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