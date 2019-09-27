package model

import (
	"fmt"
	"io"
	"sync"
)

var context *Context
var once sync.Once

// Context 执行环境
type Context struct {
	writer io.Writer
	data   map[string]interface{}
}

//Print 打印一行
func (c *Context) Print(a ...interface{}) (n int, err error) {
	text := fmt.Sprint(a...)
	n, err = c.writer.Write([]byte(text))
	return
}

//Println 打印一行
func (c *Context) Println(a ...interface{}) (n int, err error) {
	text := fmt.Sprintln(a...)
	n, err = c.writer.Write([]byte(text))
	return
}

//Printf 根据格式字符串打印
func (c *Context) Printf(format string, a ...interface{}) (n int, err error) {
	text := fmt.Sprintf(format, a...)
	n, err = c.writer.Write([]byte(text))
	return
}

//PrintErr 打印错误
func (c *Context) PrintErr(uerr error) (n int, err error) {
	text := fmt.Sprintf("Error:%v\n", uerr)
	n, err = c.writer.Write([]byte(text))
	return
}

//GetWriter 获取Writer
func (c *Context) GetWriter() io.Writer {
	return c.writer
}

// GetContext 创建一个单例的Context
func GetContext(writer io.Writer) *Context {
	once.Do(func() {
		data := make(map[string]interface{}, 0)
		context = &Context{writer, data}
	})
	return context
}
