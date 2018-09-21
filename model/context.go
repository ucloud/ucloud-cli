package model

import (
	"io"
	"sync"

	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/trace"
)

var context *Context
var once sync.Once

// Context 执行环境
type Context struct {
	writer    io.Writer
	TraceInfo *trace.DasTraceInfo
}

// GetContext 创建一个单例的Context
func GetContext(writer io.Writer, client *sdk.Client) *Context {
	once.Do(func() {
		traceInfo := client.Tracer.DasTraceInfo
		context = &Context{writer, &traceInfo}
	})
	return context
}
