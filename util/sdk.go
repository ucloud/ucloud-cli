package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

const ConfigPath = ".ucloud"

//GetHomePath 获取家目录
func GetHomePath() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

//MosaicString 对字符串敏感部分打马赛克 如公钥私钥
func MosaicString(s string, beginChars, lastChars int) string {
	r := len(s) - lastChars - beginChars
	if r > 0 {
		return s[:beginChars] + strings.Repeat("*", r) + s[(r+beginChars):]
	}
	return strings.Repeat("*", len(s))
}

//AppendToFile 添加到文件中
func AppendToFile(name string, content string) error {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("\n%s\n", content))
	return err
}

//LineInFile 检查某一行是否在某文件中
func LineInFile(fileName string, lookFor string) bool {
	f, err := os.Open(fileName)
	if err != nil {
		return false
	}
	defer f.Close()
	r := bufio.NewReader(f)
	prefix := []byte{}
	for {
		line, isPrefix, err := r.ReadLine()
		if err == io.EOF {
			return false
		}
		if err != nil {
			return false
		}
		if isPrefix {
			prefix = append(prefix, line...)
			continue
		}
		line = append(prefix, line...)
		if string(line) == lookFor {
			return true
		}
		prefix = prefix[:0]
	}
}

func GetConfigPath() string {
	path := GetHomePath() + "/" + ConfigPath
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		if err != nil {
			panic(err)
		}
	}
	return path
}
