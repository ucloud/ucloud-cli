package base

import (
	"fmt"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
)

//Logger 日志
var logger *log.Logger
var mu sync.Mutex
var out = Cxt.GetWriter()

func initConfigDir() {
	if _, err := os.Stat(GetLogFileDir()); os.IsNotExist(err) {
		err := os.MkdirAll(GetLogFileDir(), 0755)
		if err != nil {
			panic(err)
		}
	}
}
func initLog() error {
	file, err := os.OpenFile(GetLogFilePath(), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open log file failed: %v", err)
	}
	logger = log.New()
	logger.SetNoLock()
	logger.AddHook(NewLogRotateHook(file))
	logger.SetOutput(file)
	LogInfo(fmt.Sprintf("command: %s", strings.Join(os.Args, " ")))
	return nil
}

//GetLogFileDir 获取日志文件路径
func GetLogFileDir() string {
	return GetHomePath() + fmt.Sprintf("/%s", ConfigPath)
}

//GetLogFilePath 获取日志文件路径
func GetLogFilePath() string {
	return GetHomePath() + fmt.Sprintf("/%s/cli.log", ConfigPath)
}

//LogInfo 记录日志
func LogInfo(logs ...string) {
	mu.Lock()
	defer mu.Unlock()
	for _, line := range logs {
		logger.Info(line)
	}
}

//LogError 记录日志
func LogError(logs ...string) {
	for _, line := range logs {
		logger.Error(line)
		fmt.Fprintln(out, line)
	}
}

//LogRotateHook rotate log file
type LogRotateHook struct {
	MaxSize int64
	Cut     float32
	LogFile *os.File
	mux     sync.Mutex
}

//Levels fires hook
func (hook *LogRotateHook) Levels() []log.Level {
	return log.AllLevels
}

//Fire do someting when hook is triggered
func (hook *LogRotateHook) Fire(entry *log.Entry) error {
	hook.mux.Lock()
	defer hook.mux.Unlock()
	info, err := hook.LogFile.Stat()
	if err != nil {
		return err
	}

	if info.Size() <= hook.MaxSize {
		return nil
	}
	hook.LogFile.Sync()
	offset := int64(float32(hook.MaxSize) * hook.Cut)
	buf := make([]byte, info.Size()-offset)
	_, err = hook.LogFile.ReadAt(buf, offset)
	if err != nil {
		return err
	}

	nfile, err := os.Create(GetLogFilePath() + ".tmp")
	if err != nil {
		return err
	}
	nfile.Write(buf)
	nfile.Close()

	err = os.Rename(GetLogFilePath()+".tmp", GetLogFilePath())
	if err != nil {
		return err
	}

	mfile, err := os.OpenFile(GetLogFilePath(), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("open log file failed: ", err)
		return err
	}
	entry.Logger.SetOutput(mfile)
	initLog()
	return nil
}

//NewLogRotateHook create a LogRotateHook
func NewLogRotateHook(file *os.File) *LogRotateHook {
	return &LogRotateHook{
		MaxSize: 1024 * 1024, //1MB
		Cut:     0.2,
		LogFile: file,
	}
}

//ToQueryMap tranform request to map
func ToQueryMap(req request.Common) map[string]string {
	reqMap, err := request.ToQueryMap(req)
	if err != nil {
		return nil
	}
	delete(reqMap, "Password")
	return reqMap
}
