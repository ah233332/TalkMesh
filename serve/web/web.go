package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

// 编译器指令
//
//go:embed static
var staticFS embed.FS

//相当于整了一个文件系统，里面包含这static目录下面所有文件内容
//把整个 static/ 文件夹压进 exe 里，并提供一个只读的 Open/ReadFile 接口

func Start(addr string) error {
	//静态文件：embed 的根是项目目录，切到 static/ 后按 /static/xxx 访问
	//sub相当于创建了一个新的文件系统视图，只能往子文件目录里面进
	//s.Sub(fsys, dir) 的语义是“把 dir 当成新的根目录”
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return err
	}

	//创建一个新的路由器
	mux := http.NewServeMux()
	//http.FS(sub) 就是转接头，把 fs.FS 适配成 http.FileSystem
	//http.FileServer 创建一个“按路径找文件并返回”的通用 handler
	//http.StripPrefix 请求进来先去掉 /static/ 前缀，再交给文件服务器
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))

	//handle 给"会处理请求的对象"
	//handlefunc 直接给"函数"

	//路径 函数
	//创建路由表单，设置路径与对应函数，把路径和函数绑在一起
	//这里的函数不是普通函数，是经过handlefunc包装过的函数，是handler接口值
	//匹配规则是最长前缀优先
	mux.HandleFunc("/", index)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/register", register)
	mux.HandleFunc("/chat", chat)
	mux.HandleFunc("/api/messages", apiMessages)
	mux.HandleFunc("/sendall", sendall)
	mux.HandleFunc("/sendone", sendone)
	mux.HandleFunc("/logout", logout)

	//循环监听和服务
	// 只有出错（端口被占用、监听失败）才返回
	return http.ListenAndServe(addr, mux)

	//listenandserve第二个变量如果是nil，代表使用全局默认路由器
	//可能会与第三方其他库撞车
	//好处是省事，坏处是全局共享
}

// w响应文本，把响应文本的组装封装成一个接口，把响应信息按http格式返回浏览器
// r请求文本，把浏览器的请求解析成一个结构体方法，可以在go中解析操作浏览器的请求文本
func index(w http.ResponseWriter, r *http.Request) {
	//从 embed 的文件系统里读一个文件，设好类型，返回给浏览器
	http.ServeFileFS(w, r, staticFS, "static/index.html")
}

// 登录
func login(w http.ResponseWriter, r *http.Request) {
	//0. 读浏览器的表单信息
	r.ParseForm()
	id, err := strconv.Atoi(r.FormValue("userId"))
	if err != nil {
		fmt.Fprint(w, "用户ID必须是数字")
		return
	}
	pwd := r.FormValue("userPwd")

	// 1. 连服务端
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8899", 3*time.Second)
	if err != nil {
		fmt.Fprintf(w, "连接失败: %v", err)
		return
	}

	// 2. 组 LoginMes
	var mes message.Message
	mes.Type = message.LoginMesType

	loginMes := message.LoginMes{UserId: id, UserPwd: pwd}
	data, err := json.Marshal(loginMes)
	if err != nil {
		fmt.Fprintf(w, "序列化失败: %v", err)
		conn.Close()
		return
	}
	mes.Data = string(data)

	pkg, err := json.Marshal(mes)
	if err != nil {
		fmt.Fprintf(w, "序列化失败: %v", err)
		conn.Close()
		return
	}

	// 3. 发送并读响应（5 秒超时）
	tf := utils.Transfer{Conn: conn, Timeout: 5 * time.Second}
	if err := tf.WritePkg(pkg); err != nil {
		fmt.Fprintf(w, "发送失败: %v", err)
		conn.Close()
		return
	}

	resMes, err := tf.ReadPkg()
	if err != nil {
		fmt.Fprintf(w, "读取响应失败: %v", err)
		conn.Close()
		return
	}

	var res message.LoginResMes
	if err := json.Unmarshal([]byte(resMes.Data), &res); err != nil {
		fmt.Fprintf(w, "解析响应失败: %v", err)
		conn.Close()
		return
	}

	if res.Code != 200 {
		conn.Close() // 失败才关
		fmt.Fprintf(w, "登录失败: %s", res.Error)
		return
	}

	userName := ""
	for _, u := range res.Users {
		if u.UserId == id {
			userName = u.UserName
			break
		}
	}

	users := make(map[int]*message.User, len(res.Users))
	for i := range res.Users {
		u := res.Users[i]
		users[u.UserId] = &u
	}

	ss := &Session{
		Conn:     conn,
		UserId:   id,
		UserName: userName,

		Users: users,
	}

	//登录成功
	//设置cookie专属的校验码
	//分别存到session表单与浏览器中
	sid := newSessionID()
	if sid == "" {
		conn.Close()
		fmt.Fprint(w, "生成会话失败")
		return
	}

	addSession(sid, ss)

	//开启读取服务端发送的信息
	go ss.readloop()

	//开启心跳协程
	go ss.sendping()

	http.SetCookie(w, &http.Cookie{Name: "sid", Value: sid, Path: "/"})
	//返回一个“请去这个地址”的响应,浏览器收到后自动再请求一次
	//它不执行目标页面的代码
	http.Redirect(w, r, "/chat", http.StatusSeeOther)
}

// 注册
func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFileFS(w, r, staticFS, "static/register.html")
		return
	}

	r.ParseForm()
	id, err := strconv.Atoi(r.FormValue("userId"))
	name := r.FormValue("userName")
	pwd := r.FormValue("userPwd")
	if err != nil || id <= 0 || name == "" || pwd == "" {
		fmt.Fprint(w, "用户ID必须是正整数，用户名和密码不能为空，<a href=\"/register\">返回</a>")
		return
	}

	conn, err := net.DialTimeout("tcp", "127.0.0.1:8899", 3*time.Second)
	if err != nil {
		fmt.Fprintf(w, "连接失败: %v", err)
		return
	}
	defer conn.Close()

	regMes := message.RegisterMes{
		User: message.User{
			UserId:   id,
			UserName: name,
			UserPwd:  pwd,
		},
	}
	data, err := json.Marshal(regMes)
	if err != nil {
		fmt.Fprintf(w, "序列化失败: %v", err)
		return
	}

	mes := message.Message{Type: message.RegisterMesType, Data: string(data)}
	pkg, err := json.Marshal(mes)
	if err != nil {
		fmt.Fprintf(w, "序列化失败: %v", err)
		return
	}

	tf := utils.Transfer{Conn: conn, Timeout: 5 * time.Second}
	if err := tf.WritePkg(pkg); err != nil {
		fmt.Fprintf(w, "发送失败: %v", err)
		return
	}

	resMes, err := tf.ReadPkg()
	if err != nil {
		fmt.Fprintf(w, "读取响应失败: %v", err)
		return
	}

	var res message.RegisterResMes
	if err := json.Unmarshal([]byte(resMes.Data), &res); err != nil {
		fmt.Fprintf(w, "解析响应失败: %v", err)
		return
	}

	if res.Code != 200 {
		fmt.Fprintf(w, "注册失败: %s，<a href=\"/register\">返回</a>", res.Error)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// 聊天
func chat(w http.ResponseWriter, r *http.Request) {
	//获取cookie
	cookie, err := r.Cookie("sid")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	//验证cookie
	if _, ok := getSession(cookie.Value); !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.ServeFileFS(w, r, staticFS, "static/chat.html")
}

// msgView 是给前端渲染用的消息结构
type msgView struct {
	Type     string `json:"type,omitempty"` // group / private / notify / result
	UserId   int    `json:"userId,omitempty"`
	UserName string `json:"userName,omitempty"`
	Content  string `json:"content,omitempty"`
	Text     string `json:"text,omitempty"`
	Time     string `json:"time"`
}

type messagesResp struct {
	Me       msgView   `json:"me"`
	Messages []msgView `json:"messages"`
	Users    []msgView `json:"users"`
}

// statusText 把私聊结果码翻译成中文
func statusText(code int) string {
	switch code {
	case 200:
		return "目标用户在线，发送成功"
	case 201:
		return "目标用户离线，成功发送离线消息"
	case 400:
		return "目标用户不在线"
	case 401:
		return "用户未登录"
	case 402:
		return "发送对象错误"
	case 403:
		return "发送者身份不一致"
	case 405:
		return "目标用户不存在"
	default:
		return "发送失败"
	}
}

// viewOf 把协议消息转换成前端好渲染的结构
func (ss *Session) viewOf(mes message.Message, at time.Time) (msgView, bool) {
	switch mes.Type {
	case message.SmsTransferMesType:
		//群聊信息
		var m message.SmsTransferMes
		if err := json.Unmarshal([]byte(mes.Data), &m); err != nil {
			return msgView{}, false
		}
		return msgView{Type: "group", UserId: m.UserId, UserName: m.UserName, Content: m.Content, Time: at.Format("15:04")}, true
	case message.SmsToUserTransferMesType:
		//私聊信息
		var m message.SmsToUserTransferMes
		if err := json.Unmarshal([]byte(mes.Data), &m); err != nil {
			return msgView{}, false
		}
		return msgView{Type: "private", UserId: m.UserId, UserName: m.UserName, Content: m.Content, Time: at.Format("15:04")}, true
	case message.NotifyUserStatusMesType:
		//状态变化信息
		var m message.NotifyUserStatusMes
		if err := json.Unmarshal([]byte(mes.Data), &m); err != nil {
			return msgView{}, false
		}
		text := "用户状态变化"
		switch m.Status {
		case message.UserOnline:
			text = "上线了"
		case message.UserOffline:
			text = "离线了"
		case message.UserBusyStatus:
			text = "忙碌"
		}
		return msgView{Type: "notify", UserId: m.UserId, UserName: m.UserName, Text: text, Time: at.Format("15:04")}, true
	case message.SmsToUserResMesType:
		//私聊响应信息
		var m message.SmsToUserResMes
		if err := json.Unmarshal([]byte(mes.Data), &m); err != nil {
			return msgView{}, false
		}
		return msgView{Type: "result", Text: statusText(m.Code), Time: at.Format("15:04")}, true
	default:
		return msgView{}, false
	}
}

// apiMessages 给前端轮询用，返回最近的消息
func apiMessages(w http.ResponseWriter, r *http.Request) {
	//获取cookie
	cookie, err := r.Cookie("sid")
	if err != nil {
		http.Error(w, "未登录", http.StatusUnauthorized)
		return
	}
	//验证cookie
	ss, ok := getSession(cookie.Value)
	if !ok {
		http.Error(w, "未登录", http.StatusUnauthorized)
		return
	}

	//遍历传输消息列表
	msgs := ss.snapshot()
	views := make([]msgView, 0, len(msgs))
	for _, mes := range msgs {
		if v, ok := ss.viewOf(mes.Meg, mes.At); ok {
			views = append(views, v)
		}
	}

	//遍历在线用户列表
	users := ss.snapshotUsers()
	sort.Slice(users, func(i, j int) bool {
		return users[i].UserId < users[j].UserId
	})
	userViews := make([]msgView, 0, len(users))
	for _, v := range users {
		userViews = append(userViews, msgView{
			UserId:   v.UserId,
			UserName: v.UserName,
		})
	}

	resp := messagesResp{
		Me:       msgView{UserId: ss.UserId, UserName: ss.UserName},
		Messages: views,
		Users:    userViews,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	data, _ := json.Marshal(resp)
	w.Write(data)
}

// 发送群聊信息
func sendall(w http.ResponseWriter, r *http.Request) {
	//获取cookie
	cookie, err := r.Cookie("sid")
	if err != nil {
		fmt.Println("get cookie err=", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//验证cookie
	session, ok := getSession(cookie.Value)
	if !ok {
		fmt.Println("check cookie err")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//读取浏览器上的信息
	r.ParseForm()
	content := r.FormValue("content")
	if content == "" {
		http.Redirect(w, r, "/chat", http.StatusSeeOther)
		return
	}

	//封装群发信息
	var mes message.Message
	mes.Type = message.SmsMessageType
	groupmes := message.SmsMessage{
		UserId:     session.UserId,
		UserName:   session.UserName,
		UserStatus: message.UserOnline,
		Content:    content,
	}
	data, err := json.Marshal(groupmes)
	if err != nil {
		fmt.Println("marshal groupmes err=", err)
		http.Redirect(w, r, "/chat", http.StatusSeeOther)
		return
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		http.Redirect(w, r, "/chat", http.StatusSeeOther)
		return
	}
	//此时data就是可直接发送的群发请求

	tf := utils.Transfer{
		Conn: session.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("writePKG mes err=", err)
		http.Redirect(w, r, "/chat", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/chat", http.StatusSeeOther)

}

// 发送私聊信息
func sendone(w http.ResponseWriter, r *http.Request) {
	//获取cookie
	cookie, err := r.Cookie("sid")
	if err != nil {
		fmt.Println("get cookie err=", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//验证cookie
	session, ok := getSession(cookie.Value)
	if !ok {
		fmt.Println("check cookie err")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//从浏览器上获取信息
	r.ParseForm()
	toUserId, err := strconv.Atoi(r.FormValue("toUserId"))
	if err != nil {
		fmt.Println("atio toUserID err=", err)
		http.Redirect(w, r, "/chat", http.StatusSeeOther)
		return
	}
	content := r.FormValue("content")
	if content == "" {
		http.Redirect(w, r, "/chat", http.StatusSeeOther)
		return
	}

	//封装私聊信息
	var mes message.Message
	mes.Type = message.SmsToUserMesType
	toOneMes := message.SmsToUserMes{
		UserId:     session.UserId,
		UserName:   session.UserName,
		UserStatus: message.UserOnline,

		ToUserId: toUserId,
		Content:  content,
	}

	data, err := json.Marshal(toOneMes)
	if err != nil {
		fmt.Println("marshal toOneMes err=", err)
		http.Redirect(w, r, "/chat", http.StatusSeeOther)
		return
	}

	mes.Data = string(data)

	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("marshal mes err=", err)
		http.Redirect(w, r, "/chat", http.StatusSeeOther)
		return
	}
	//此时data就是可直接发送的群发请求

	tf := utils.Transfer{
		Conn: session.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("writePKG mes err=", err)
		http.Redirect(w, r, "/chat", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/chat", http.StatusSeeOther)
}

// 退出
func logout(w http.ResponseWriter, r *http.Request) {
	//获取cookie
	cookie, err := r.Cookie("sid")
	if err != nil {
		fmt.Println("get cookie err=", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//验证cookie
	_, ok := getSession(cookie.Value)
	if !ok {
		fmt.Println("check cookie err")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//删除map记录，断开conn连接
	delSession(cookie.Value)

	//服务端让会话失效
	//同时浏览器也要把旧凭证清理掉

	//清除cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "sid",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
