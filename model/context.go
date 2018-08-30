package model

import (
	"fmt"
	"io"
	"os"
	"sync"
)

var context *Context
var once sync.Once

// Context 执行环境
type Context struct {
	writer io.Writer
}

// Println 在当前执行环境打印一行
func (p *Context) Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(p.writer, a...)
}

// Print 在当前执行环境打印一串
func (p *Context) Print(a ...interface{}) (n int, err error) {
	return fmt.Fprint(p.writer, a...)
}

// GetContext 创建一个单例的Context
func GetContext() *Context {
	once.Do(func() {
		context = &Context{os.Stdout}
	})
	return context
}
