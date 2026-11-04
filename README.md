# TalkMesh

基于 Go 的 TCP 即时通讯学习项目。目标是把一条完整的 IM 链路走通：
连接 → 协议编解码 → 登录/注册 → 在线用户 → 群聊/私聊 → 心跳 → 离线消息。

当前状态：**登录、注册、在线用户（上线/下线/心跳保活）、群聊与私聊收发均已打通**。客户端可以注册、登录、查看在线用户，在菜单 2 发群聊、菜单 3 发私聊；某人断开、或 30 秒内没有任何心跳/消息时，在线列表会自动移除他。离线消息、网页界面、消息可靠性尚未实现。

## 已实现

- 客户端与服务端 TCP 长连接，服务端为每个连接启动一个 goroutine 处理
- 自定义协议：4 字节大端长度头 + JSON 消息体，用长度头切分消息边界。收发使用 `io.ReadFull` 读满包头/包体，并校验包长不超过缓冲区（8096 字节），超长或半包会返回错误并关闭连接
- 按 `Message.Type` 路由到对应业务处理函数，服务端目前处理 `LoginMes` / `RegisterMes` / `SmsMessage`
- 登录流程：客户端组装 `LoginMes` → 服务端查询 Redis → 返回 `LoginResMes{Code, Error, UserIds}`
- 注册消息结构：已定义 `RegisterMes` / `RegisterResMes`
- 客户端注册流程：控制台输入用户信息 → 本地校验（ID 为正、密码和用户名非空）→ 发送 `RegisterMes` → 读取 `RegisterResMes`
- 服务端注册流程：反序列化 `RegisterMes` → 参数校验（不合法返回 401）→ `UserDao.Register` 查重并写入 Redis `users` 哈希 → 返回 `RegisterResMes`
- 注册结果码：200 成功 / 400 用户 ID 已存在 / 401 参数不合法 / 505 其他错误
- 用户 JSON 统一用 `json.Marshal` 序列化后写入，密码或用户名包含 `"`、`\` 等字符也不会写坏数据
- 客户端按 `main` / `process` 分层，登录成功后由独立 goroutine 持续接收服务端消息
- 登录后的控制台菜单：在线用户列表、群发消息可用，消息列表仍为占位
- 在线用户维护：服务端用 `UserMgr` 记录在线连接，登录成功后向其他在线用户推送 `NotifyUserStatusMes`，客户端收到后更新本地在线列表
- 在线用户表的并发安全：`UserMgr` 用 `sync.RWMutex` 保护，增删查都加锁，遍历前先 `Snapshot` 复制一份名单，网络发送在锁外进行
- 下线清理：连接结束时按 `UserId` 从在线表删除，并向其他在线用户广播 `UserOffline`；客户端收到后从本地列表移除。未登录过的连接（`UserId` 为 0）不参与广播
- 心跳保活：客户端每 10 秒发 `PingMes`，服务端回 `PongMes`；两端读超时都设为 30 秒，超时未收到任何包就判定对端失联，服务端据此关闭连接并触发下线清理
- 读超时：`Transfer` 支持按次设置读超时（`Timeout`，0 表示不超时）；登录/注册 5 秒等不到响应会报错退出，登录后的收消息循环不设超时，每次读取都会重设连接的 deadline，避免沿用上一阶段的超时
- 群聊收发：客户端把 `SmsMessage`（`userId` / `userName` / `status` / `content`）发给服务端，服务端把外层类型改成 `SmsTransferMes` 后转发给所有在线用户（含发送者），客户端收到后打印「用户名(id)发送了一条消息:内容」
- 私聊：客户端发 `SmsToUserMes`（`userId` / `userName` / `status` / `toUserId` / `content`），服务端按 `toUserId` 查在线表，找到就把类型改成 `SmsToUserTransferMes` 转发给目标，并给发送者回 `SmsToUserResMes`（200 成功 / 400 对方不在线 / 505 其他错误）；客户端收到私聊显示「用户名(id)给你私发了一条消息:内容」
- 数据访问层：`UserDao` 使用 redigo 连接池读取 Redis 的 `users` 哈希，校验用户密码；注册时做重名校验并写入

## 项目结构

```text
TalkMesh/
├── client/
│   ├── main/
│   │   └── main.go          # 控制台入口与登录/注册菜单
│   ├── model/
│   │   └── curUser.go       # 当前登录用户与连接
│   └── process/
│       ├── userProcess.go   # 登录、注册：本地校验、发送请求并处理结果
│       ├── userMgr.go       # 客户端在线用户缓存与状态更新
│       ├── serve.go         # 登录后菜单、持续接收服务端消息
│       ├── smsProcess.go    # 发送群聊消息
│       ├── smsMgr.go        # 接收并显示群聊消息
│       └── pingProcess.go   # 心跳：定时发送 PingMes
├── common/
│   ├── message/
│   │   ├── message.go       # 协议：Message / LoginMes / LoginResMes / RegisterMes / RegisterResMes / SmsMessage / SmsTransferMes / PingMes / PongMes
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
│   │   ├── smsProcess.go    # 群聊消息转发
│   │   └── pingProcess.go   # 心跳：回 PongMes
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

一个群聊包的完整样子：

```json
{"type":"SmsMessage","data":"{\"userId\":114514,\"userName\":\"测试用户\",\"status\":1,\"content\":\"大家好\"}"}
```

服务端转发时只改外层类型（`SmsMessage` → `SmsTransferMes`），`data` 内容不变。

已定义的消息类型：

| 类型 | 方向 | 说明 |
| --- | --- | --- |
| `LoginMes` | 客户端 → 服务端 | 登录请求，`userId` / `userPwd` / `userName` |
| `LoginResMes` | 服务端 → 客户端 | 登录结果，`code`：200 成功 / 403 密码错误 / 500 用户不存在 / 505 未知错误，`userIds` 为当前在线用户 ID 列表 |
| `RegisterMes` | 客户端 → 服务端 | 注册请求，`user` 字段包含 `userId` / `userPwd` / `userName` |
| `RegisterResMes` | 服务端 → 客户端 | 注册结果，`code`：200 成功 / 400 用户已存在 / 401 参数不合法 / 505 其他错误，`error` 为提示信息 |
| `NotifyUserStatusMes` | 服务端 → 客户端 | 用户状态通知，`userId` / `status`：1 在线 / 2 离线 / 3 忙碌 |
| `SmsMessage` | 客户端 → 服务端 | 群发消息，`userId` / `userName` / `status` / `content` |
| `SmsTransferMes` | 服务端 → 客户端 | 群发消息转发，服务端把 `mes.Type` 改成它后发给所有在线用户，字段与 `SmsMessage` 相同 |
| `SmsToUserMes` | 客户端 → 服务端 | 私聊请求，`userId` / `userName` / `status` / `toUserId` / `content` |
| `SmsToUserTransferMes` | 服务端 → 客户端 | 私聊转发给目标用户，字段与 `SmsToUserMes` 相同 |
| `SmsToUserResMes` | 服务端 → 客户端 | 私聊结果，`code`：200 成功 / 400 对方不在线 / 505 其他错误 |
| `PingMes` | 客户端 → 服务端 | 心跳请求，`timestamp` 为发送时刻的毫秒时间戳 |
| `PongMes` | 服务端 → 客户端 | 心跳响应，原样带回 `timestamp` |

## 开发计划

> 阶段顺序： 私聊 →  离线消息 → 网页界面 → 消息可靠性 → 安全 → 测试与部署。

### 协议与基础

- [x] TCP 连接、长度头 + JSON 协议、按类型路由
- [x] 客户端按 `main` / `process` 分层，登录后独立 goroutine 收消息
- [x] 修复分包读取：`io.ReadFull` + 包长上限校验
- [x] 写操作加锁：`WritePkg` 的包头/包体整体加锁，避免多个 goroutine 写同一连接时交错
- [x] 客户端增加读超时，检查 `ReadPkg` 错误，区分「没响应」和「响应解析失败」

### 登录与注册

- [x] 登录链路打通（客户端 → 服务端 → 返回结果）
- [x] 引入 redigo、初始化连接池并接入 `UserDao`
- [x] 登录改为 Redis 校验
- [x] 定义登录/注册消息结构，完成客户端注册请求发送
- [x] 实现服务端注册：处理 `RegisterMes` + Redis 写入 + 重名校验 + 参数校验（401）
- [x] 注册改用 `HSETNX`，消除并发重复注册

### 在线用户

- [x] 服务端维护在线用户表，登录后向其他在线用户推送上线通知
- [x] 客户端保存在线用户并在菜单中显示列表
- [x] 用户下线时从在线表移除（`DelOnlineUser`），并向其他在线用户推送离线通知
- [x] 客户端收到离线通知后从本地列表移除
- [x] 应用层心跳，及时检测断线并触发清理
- [x] 在线用户表加锁，避免多个连接并发读写 map
- [x] 用户状态常量从 1 开始，避免与 `User` 零值冲突

### 群聊

- [x] 客户端发送群聊消息（`SmsMessage`）
- [x] 服务端把群聊消息转发给所有在线用户（含发送者）
- [x] 客户端解析并显示收到的群聊消息（`outputGroupMes`）
- [x] 单个接收者发送失败时不影响发送者连接
- [ ] 登录返回用户名，避免群聊消息里 `userName` 为空
- [ ] 群聊内容支持空格（当前 `fmt.Scan` 只读一个词）

### 私聊

- [x] 私聊消息结构：`SmsToUserMes`（`userId` / `userName` / `status` / `toUserId` / `content`）
- [x] 服务端收到后按 `toUserId` 查在线表，找到连接就转发
- [x] 客户端菜单支持输入收件人 ID 并发送
- [x] 客户端收到私聊后单独显示，与群聊区分
- [x] 对方不在线时，给发送者回明确提示（400）

### 离线消息

- [ ] 私聊对象不在线时，把消息写入 Redis list（`offline:<userId>`）
- [ ] 登录成功后读取该用户的离线消息，逐条补发
- [ ] 补发完成后清理，避免重复投递
- [ ] 每人离线消息条数上限（`LTRIM`）
- [ ] 客户端收到离线消息后的显示（复用私聊样式或单独提示）

### 网页界面

- [ ] 同进程增加 8080 HTTP 服务，托管静态页面（`embed`）
- [ ] `POST /api/login`、`POST /api/register`
- [ ] `GET /api/users` 返回在线列表
- [ ] `POST /api/messages` 发群聊，页面轮询拉取新消息
- [ ] 登录页与聊天页（在线列表 / 消息区 / 输入框）
- [ ] （可选）网页用户与终端用户共用在线表，需要 Session 抽象

### 消息可靠性

- [ ] 每条消息生成唯一 ID
- [ ] 接收方回 ACK，发送方超时重发
- [ ] 接收方按 ID 去重
- [ ] 断线重连后补拉未确认的消息

### 安全

- [ ] 服务端用连接自身的 `UserId` 作为发送者，忽略消息里伪造的身份
- [ ] 密码改为 bcrypt 哈希存储
- [ ] 登录/注册接口限流
- [ ] （可选）TLS 加密传输

### 测试与部署

- [ ] `ReadPkg` 的半包、超长包单元测试（`net.Pipe`）
- [ ] `UserDao` 注册重名、登录校验的单元测试
- [ ] 多连接并发压测，记录连接数与消息吞吐
- [ ] 结构化日志、配置化（端口 / Redis / 超时）
- [ ] 优雅退出（关闭监听、清理连接）
- [ ] 部署（Dockerfile 或 systemd）

## 说明

本项目用于学习 Go 网络编程、协议设计与并发处理，尚未达到生产级聊天系统的标准。
