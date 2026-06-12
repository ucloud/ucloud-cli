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
// 结果通过返回的 channel（缓冲1）投递；首个回调投递后触发 server 关闭（sync.Once 保证只投递一次）。
func startCallbackServer(ln net.Listener, expectState string) (*http.Server, <-chan callbackResult) {
	ch := make(chan callbackResult, 1)
	var once sync.Once

	srv := &http.Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/authorization", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		var res callbackResult

		if e := q.Get("error"); e != "" {
			if e == "access_denied" {
				res.err = fmt.Errorf("authorization was denied in the browser. Run 'ucloud auth login' to try again")
			} else {
				res.err = fmt.Errorf("oauth server returned error %q. Run 'ucloud auth login' to try again", e)
			}
		} else if code := q.Get("code"); code == "" {
			res.err = fmt.Errorf("callback carried no authorization code. Run 'ucloud auth login' to try again")
		} else if q.Get("state") != expectState {
			res.err = fmt.Errorf("state mismatch: the callback likely comes from a previous login attempt. Run 'ucloud auth login' again")
		} else {
			res.code = code
		}

		if res.err != nil {
			http.Error(w, "Login failed. Return to the terminal for details.", http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, callbackSuccessHTML)
		}

		once.Do(func() {
			ch <- res
		})
	})
	srv.Handler = mux

	go srv.Serve(ln)
	return srv, ch
}
