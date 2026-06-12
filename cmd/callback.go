// cmd/callback.go
package cmd

import (
	"fmt"
	"net"
	"net/http"
	"sync"
)

// allocateLoopbackListener 在 127.0.0.1 上取一个内核分配的空闲端口（>=1024），返回 listener 与端口。
func allocateLoopbackListener() (net.Listener, int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, 0, fmt.Errorf("cannot open a local callback port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	return ln, port, nil
}

type callbackResult struct {
	code string
	err  error
}

const callbackSuccessHTML = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>Login successful</title></head>
<body style="font-family:sans-serif;text-align:center;margin-top:15%">
<h2>Login successful</h2>
<p>You can close this tab and return to the terminal.</p>
<p>登录成功，可关闭此页面返回终端。</p>
</body>
</html>`

// startCallbackServer 在给定 listener 上起一个临时 HTTP server，只处理 GET /authorization。
// 结果通过返回的 channel（缓冲1）投递，sync.Once 保证只投递一次。投递规则：
//   - error 参数（如 access_denied）→ 回 400 并投递错误（中止登录）；
//   - code + state 匹配 → 回成功页并投递 code；
//   - 缺 code 或 state 不匹配（本地探针/陈旧标签页等噪音请求）→ 仅回 400、不投递，
//     继续等待真正的回调（上层 loginCallbackTimeout 兜底）。
func startCallbackServer(ln net.Listener, expectState string) (*http.Server, <-chan callbackResult) {
	ch := make(chan callbackResult, 1)
	var once sync.Once

	srv := &http.Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/authorization", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if e := q.Get("error"); e != "" {
			var err error
			if e == "access_denied" {
				err = fmt.Errorf("authorization was denied in the browser. Run 'ucloud auth login' to try again")
			} else {
				err = fmt.Errorf("oauth server returned error %q. Run 'ucloud auth login' to try again", e)
			}
			http.Error(w, "Login failed. Return to the terminal for details.", http.StatusBadRequest)
			once.Do(func() {
				ch <- callbackResult{err: err}
			})
			return
		}

		code := q.Get("code")
		if code == "" || q.Get("state") != expectState {
			// 噪音请求：不消耗 once，登录继续等待真正的回调
			http.Error(w, "Login failed. Return to the terminal for details.", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, callbackSuccessHTML)
		once.Do(func() {
			ch <- callbackResult{code: code}
		})
	})
	srv.Handler = mux

	go srv.Serve(ln)
	return srv, ch
}
