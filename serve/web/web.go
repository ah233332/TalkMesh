package web

import (
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/ah233332/TalkMesh/common/message"
	"github.com/ah233332/TalkMesh/common/utils"
)

func Start(addr string) error {
	//创建一个新的路由器
	mux := http.NewServeMux()

	//路径 函数
	//创建路由表单，设置路径与对应函数，把路径和函数绑在一起
	//这里的函数不是普通函数，是经过handlefunc包装过的函数，是handler接口值
	//匹配规则是最长前缀优先
	mux.HandleFunc("/", index)
	mux.HandleFunc("/login", login)
	mux.HandleFunc("/chat", chat)
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
	fmt.Fprint(w, `<!DOCTYPE html>
  <html>
	<body>
		<h1>TalkMesh</h1>
		<form method="post" action="/login">
				<input name="userId" placeholder="用户ID">
				<input name="userPwd" type="password" placeholder="密码">
   			 <button>登录</button>
 		 </form>
 	 </body>
  </html>`)
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
	ss := &Session{
		Conn:     conn,
		UserId:   id,
		UserName: userName,
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

// 聊天
func chat(w http.ResponseWriter, r *http.Request) {
	//从请求中获取cookie
	cookie, err := r.Cookie("sid")
	if err != nil {
		fmt.Println("cookie 获取失败 err=", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//在session表单中验证cookie
	sess, ok := getSession(cookie.Value)
	if !ok {
		fmt.Println("getSessin cookie err")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	fmt.Fprintf(w, "<h1>欢迎%s</h1>", html.EscapeString(sess.UserName))

	//创建消息列表副本
	msgs := sess.snapshot()

	//处理服务器返回信息，然后打印到浏览器上
	fmt.Fprint(w, "<ul>")
	for _, mes := range msgs {
		switch mes.Type {
		case message.SmsTransferMesType:
			//群聊信息
			var groupmes message.SmsTransferMes
			err = json.Unmarshal([]byte(mes.Data), &groupmes)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "<li>%s(%d):%s</li>", html.EscapeString(groupmes.UserName),
				groupmes.UserId, html.EscapeString(groupmes.Content))
		case message.SmsToUserTransferMesType:
			//私聊信息
			var tomes message.SmsToUserTransferMes
			err = json.Unmarshal([]byte(mes.Data), &tomes)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "<li>%s(%d)给您私发了一条消息:%s</li>", html.EscapeString(tomes.UserName),
				tomes.UserId, html.EscapeString(tomes.Content))
		case message.SmsToUserResMesType:
			//发送私聊后的返回结果
			var toResmes message.SmsToUserResMes
			err = json.Unmarshal([]byte(mes.Data), &toResmes)
			if err != nil {
				continue
			}
			var res string
			switch toResmes.Code {
			case 200:
				res = "目标用户在线，发送成功"
			case 201:
				res = "目标用户离线，成功发送离线消息"
			case 400:
				res = "目标用户不在线"
			case 401:
				res = "用户未登录"
			case 402:
				res = "发送对象错误"
			case 403:
				res = "发送者身份不一致"
			case 405:
				res = "目标用户不存在"
			case 505:
				res = "未知错误"
			default:
				res = "未知中的未知错误"
			}
			fmt.Fprintf(w, "<li>%s</li>", html.EscapeString(res))
		case message.NotifyUserStatusMesType:
			//上线下线信息
			var notmes message.NotifyUserStatusMes
			err = json.Unmarshal([]byte(mes.Data), &notmes)
			if err != nil {
				continue
			}
			var statusString string
			switch notmes.Status {
			case message.UserOnline:
				statusString = "用户上线了"
			case message.UserOffline:
				statusString = "用户离线了"
			case message.UserBusyStatus:
				statusString = "用户繁忙"
			}

			fmt.Fprintf(w, "<li>%s(%d):%s</li>", html.EscapeString(notmes.UserName), notmes.UserId, html.EscapeString(statusString))
		default:
			fmt.Fprintf(w, "<li>%s</li>", html.EscapeString("未知信息"))
		}
	}
	fmt.Fprint(w, "</ul>")

	//点击发送群聊消息
	fmt.Fprint(w, `<form method="post" action="/sendall">
  <input name="content" placeholder="说点什么">
  <button>群发</button>
  </form>`)

	//点击发送私聊消息
	fmt.Fprint(w, `<form method="post" action="/sendone">
  发给: <input name="toUserId" placeholder="对方ID">
  内容: <input name="content">
  <button>私聊</button>
  </form>`)

	//退出
	fmt.Fprint(w, `<a href="/logout">退出</a>`)
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
