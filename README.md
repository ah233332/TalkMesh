# TalkMesh

基于 Go 的 TCP 即时通讯学习项目。目标是把一条完整的 IM 链路走通：
连接 → 协议编解码 → 登录/注册 → 在线用户 → 群聊/私聊 → 心跳 → 离线消息。

当前状态：**登录链路已打通**，服务端已通过 Redis 校验用户和密码；注册、在线用户和消息通信尚未实现。

## 已实现

- 客户端与服务端 TCP 长连接，服务端为每个连接启动一个 goroutine 处理
- 自定义协议：4 字节大端长度头 + JSON 消息体，用长度头切分消息边界。读取实现尚未处理半包，见「已知问题」
- 按 `Message.Type` 路由到对应业务处理函数，`RegisterMes` 已预留分支
- 登录流程：客户端组装 `LoginMes` → 服务端查询 Redis → 返回 `LoginResMes{Code, Error}`
- 客户端按 `main` / `process` 分层，登录成功后由独立 goroutine 持续接收服务端消息
- 登录后的控制台菜单骨架（在线列表、发消息、消息列表为占位）
- 数据访问层：`UserDao` 使用 redigo 连接池读取 Redis 的 `users` 哈希，并校验用户密码

## 项目结构

```text
TalkMesh/
├── client/
│   ├── main/
│   │   └── main.go          # 控制台入口与一级菜单
│   └── process/
│       ├── userProcess.go   # 登录：建连、发登录包、处理登录结果
│       ├── serve.go         # 登录后菜单、持续接收服务端消息
│       └── smsProcess.go    # 发送消息（待实现）
├── common/
│   ├── message/
│   │   └── message.go       # 协议：Message / LoginMes / LoginResMes / RegisterMes
│   └── utils/
│       └── utils.go         # ReadPkg / WritePkg：长度头 + JSON 的收发
├── serve/
│   ├── main/
│   │   ├── main.go          # 初始化数据库与 UserDao，监听 0.0.0.0:8899
│   │   ├── redis.go         # 创建 Redis 连接池
│   │   └── processor.go     # 读包 + 按消息类型路由
│   ├── process/
│   │   ├── userProcess.go   # 登录处理
│   │   └── smsProcess.go    # 消息处理（待实现）
│   └── model/
│       ├── user.go          # User 结构体
│       ├── error.go         # 业务错误（用户不存在 / 已存在 / 密码错误）
│       └── userDao.go       # 基于 Redis 的用户数据访问与登录校验
├── _png/                    # 客户端、服务端、交互结构图
├── go.mod
├── go.sum
└── README.md
```

设计过程中的结构图统一保存在 `_png/` 目录。

## 运行

依赖：Go 1.27+、Redis（默认地址 `127.0.0.1:6379`）。

启动服务端前，先在 Redis 的 `users` 哈希中准备用户数据，例如：

```bash
redis-cli HSET users 114514 '{"userId":114514,"userPwd":"qwe","userName":"测试用户"}'
```

然后启动服务端：

```bash
go run ./serve/main
```

再打开另一个终端启动客户端：

```bash
go run ./client/main
```

服务端默认监听 `0.0.0.0:8899`。

## 消息格式

每条消息由长度头和 JSON 消息体组成：

```text
4 字节大端长度 + JSON 数据
```

外层是统一的 `Message`，`Data` 字段里再放一层业务 JSON：

```go
type Message struct {
    Type string `json:"type"`
    Data string `json:"data"`
}
```

一个登录包的完整样子：

```json
{"type":"LoginMes","data":"{\"userId\":114514,\"userPwd\":\"qwe\",\"userName\":\"\"}"}
```

已定义的消息类型：

| 类型 | 方向 | 说明 |
| --- | --- | --- |
| `LoginMes` | 客户端 → 服务端 | 登录请求，`userId` / `userPwd` / `userName` |
| `LoginResMes` | 服务端 → 客户端 | 登录结果，`code`：200 成功 / 403 密码错误 / 500 用户不存在 / 505 未知错误，`error` 为提示信息 |
| `RegisterMes` | 客户端 → 服务端 | 注册请求，结构尚未定义 |

## 已知问题

1. `ReadPkg` 直接调用 `Conn.Read`，不保证一次读满包头/包体，遇到半包会出错，应改用 `io.ReadFull`。
2. 密码目前明文传输并以明文存储，后续增加哈希存储与安全传输。
3. 缓冲区固定 8096 字节，超过长度的包会切片越界。
4. 没有心跳，客户端拔网线时服务端无法及时感知下线。
5. 错误处理以 `fmt.Println` 为主，没有日志级别和统一错误返回。
6. 没有单元测试。

## 开发计划

- [x] TCP 连接、长度头 + JSON 协议、按类型路由
- [x] 登录链路打通（客户端 → 服务端 → 返回结果）
- [x] 客户端按 `main` / `process` 分层，登录后独立 goroutine 收消息
- [x] 引入 redigo、初始化连接池并接入 `UserDao`
- [x] 登录改为 Redis 校验
- [ ] 修复分包读取：`io.ReadFull` + 包长上限校验
- [ ] 实现注册：`RegisterMes` 结构 + Redis `hset` 写入 + 重名校验
- [ ] 维护在线用户列表并支持群聊广播
- [ ] 实现一对一聊天（按用户 ID 路由）
- [ ] 应用层心跳与断线清理
- [ ] 离线消息（Redis list，登录后补发）
- [ ] 单元测试、结构化日志、连接压测

## 说明

本项目用于学习 Go 网络编程、协议设计与并发处理，尚未达到生产级聊天系统的标准。
