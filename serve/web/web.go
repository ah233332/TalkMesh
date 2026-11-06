package web

import (
	"fmt"
	"net/http"
)

func Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", index)
	mux.HandleFunc("/login", login)
	return http.ListenAndServe(addr, mux)
}

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

func login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Fprintf(w, "收到 id=%s, pwd=%s",
		r.FormValue("userId"), r.FormValue("userPwd"))
}
