# TalkMesh

基于 Go 的 TCP 即时通讯学习项目。目标是把一条完整的 IM 链路走通：
连接 → 协议编解码 → 登录/注册 → 在线用户 → 群聊/私聊 → 心跳 → 离线消息。

当前状态：**登录、注册链路均已打通，在线用户列表可用**。客户端可以注册新用户、用注册的 ID 登录，登录后能看到当前在线用户；发消息、群聊等仍为占位。

## 已实现

- 客户端与服务端 TCP 长连接，服务端为每个连接启动一个 goroutine 处理
- 自定义协议：4 字节大端长度头 + JSON 消息体，用长度头切分消息边界。读取实现尚未处理半包
- 按 `Message.Type` 路由到对应业务处理函数，服务端目前处理 `LoginMes` 与 `RegisterMes`
- 登录流程：客户端组装 `LoginMes` → 服务端查询 Redis → 返回 `LoginResMes{Code, Error, UserIds}`
- 注册消息结构：已定义 `RegisterMes` / `RegisterResMes`
- 客户端注册流程：控制台输入用户信息 → 本地校验（ID 为正、密码和用户名非空）→ 发送 `RegisterMes` → 读取 `RegisterResMes`
- 服务端注册流程：反序列化 `RegisterMes` → 参数校验（不合法返回 401）→ `UserDao.Register` 查重并写入 Redis `users` 哈希 → 返回 `RegisterResMes`
- 注册结果码：200 成功 / 400 用户 ID 已存在 / 401 参数不合法 / 505 其他错误
- 用户 JSON 统一用 `json.Marshal` 序列化后写入，密码或用户名包含 `"`、`\` 等字符也不会写坏数据
- 客户端按 `main` / `process` 分层，登录成功后由独立 goroutine 持续接收服务端消息
- 登录后的控制台菜单：在线用户列表已可用，发消息、消息列表仍为占位
- 在线用户维护：服务端用 `UserMgr` 记录在线连接，登录成功后向其他在线用户推送 `NotifyUserStatusMes`，客户端收到后更新本地在线列表
- 数据访问层：`UserDao` 使用 redigo 连接池读取 Redis 的 `users` 哈希，校验用户密码；注册时做重名校验并写入

## 项目结构

```text
TalkMesh/
├── client/
│   ├── main/
│   │   └── main.go          # 控制台入口与登录/注册菜单
│   └── process/
│       ├── userProcess.go   # 登录、注册：本地校验、发送请求并处理结果
│       ├── userMgr.go       # 客户端在线用户缓存与状态更新
│       ├── serve.go         # 登录后菜单、持续接收服务端消息
│       └── smsProcess.go    # 发送消息（待实现）
├── common/
│   ├── message/
│   │   ├── message.go       # 协议：Message / LoginMes / LoginResMes / RegisterMes / RegisterResMes
│   │   └── user.go          # User 结构体（客户端、服务端共用的用户模型）
│   └── utils/
│       └── utils.go         # ReadPkg / WritePkg：长度头 + JSON 的收发
├── serve/
│   ├── main/
│   │   ├── main.go          # 初始化数据库与 UserDao，监听 127.0.0.1:8899
│   │   ├── redis.go         # 创建 Redis 连接池
│   │   └── processor.go     # 读包 + 按消息类型路由
│   ├── process/
│   │   ├── userProcess.go   # 登录、注册处理
│   │   ├── userMgr.go       # 服务端在线用户表与上线通知
│   │   └── smsProcess.go    # 消息处理（待实现）
│   └── model/
│       ├── error.go         # 业务错误（用户不存在 / 已存在 / 密码错误）
│       └── userDao.go       # 基于 Redis 的用户数据访问：登录校验、注册写入
├── _png/                    # 客户端、服务端、交互结构图
├── go.mod
├── go.sum
└── README.md
```

设计过程中的结构图统一保存在 `_png/` 目录。

## 运行

依赖：Go 1.27+、Redis（默认地址 `127.0.0.1:6379`）。

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

一个注册包的完整样子：

```json
{"type":"RegisterMes","data":"{\"user\":{\"userId\":114515,\"userPwd\":\"123456\",\"userName\":\"测试用户\"}}"}
```

已定义的消息类型：

| 类型 | 方向 | 说明 |
| --- | --- | --- |
| `LoginMes` | 客户端 → 服务端 | 登录请求，`userId` / `userPwd` / `userName` |
| `LoginResMes` | 服务端 → 客户端 | 登录结果，`code`：200 成功 / 403 密码错误 / 500 用户不存在 / 505 未知错误，`userIds` 为当前在线用户 ID 列表 |
| `RegisterMes` | 客户端 → 服务端 | 注册请求，`user` 字段包含 `userId` / `userPwd` / `userName` |
| `RegisterResMes` | 服务端 → 客户端 | 注册结果，`code`：200 成功 / 400 用户已存在 / 401 参数不合法 / 505 其他错误，`error` 为提示信息 |
| `NotifyUserStatusMes` | 服务端 → 客户端 | 用户状态通知，`userId` / `status`：0 在线 / 1 离线 / 2 忙碌 |

## 开发计划

- [x] TCP 连接、长度头 + JSON 协议、按类型路由
- [x] 登录链路打通（客户端 → 服务端 → 返回结果）
- [x] 客户端按 `main` / `process` 分层，登录后独立 goroutine 收消息
- [x] 引入 redigo、初始化连接池并接入 `UserDao`
- [x] 登录改为 Redis 校验
- [x] 定义登录/注册消息结构，完成客户端注册请求发送
- [x] 实现服务端注册：处理 `RegisterMes` + Redis 写入 + 重名校验 + 参数校验（401）
- [ ] 修复分包读取：`io.ReadFull` + 包长上限校验
- [ ] 注册改用 `HSETNX`，消除并发重复注册
- [ ] 客户端增加读超时，检查 `ReadPkg` 错误，区分「没响应」和「响应解析失败」
- [x] 维护在线用户列表，登录后向其他在线用户推送上线通知
- [ ] 群聊广播
- [ ] 实现一对一聊天（按用户 ID 路由）
- [ ] 应用层心跳与断线清理
- [ ] 离线消息（Redis list，登录后补发）
- [ ] 单元测试、结构化日志、连接压测

## 说明

本项目用于学习 Go 网络编程、协议设计与并发处理，尚未达到生产级聊天系统的标准。
