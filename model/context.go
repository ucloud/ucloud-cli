package model

import (
	"fmt"
	"io"
	"sync"

	"github.com/ucloud/ucloud-sdk-go/sdk"
)

var context *Context
var once sync.Once

// Context 执行环境
type Context struct {
	writer       io.Writer
	clientConfig *sdk.ClientConfig
}

// Println 在当前执行环境打印一行
func (p *Context) Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(p.writer, a...)
}

// Print 在当前执行环境打印一串
func (p *Context) Print(a ...interface{}) (n int, err error) {
	return fmt.Fprint(p.writer, a...)
}

//AppendError 添加上报的错误
func (p *Context) AppendError(err error) {
	tracerData := p.clientConfig.TracerData
	errorStr, ok := tracerData["error"].(string)
	if ok {
		tracerData["error"] = errorStr + "->" + err.Error()
	} else {
		tracerData["error"] = err.Error()
	}
}

// GetContext 创建一个单例的Context
func GetContext(writer io.Writer, clientConfig *sdk.ClientConfig) *Context {
	once.Do(func() {
		context = &Context{writer, clientConfig}
	})
	return context
}
