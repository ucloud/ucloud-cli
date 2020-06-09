package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"
	"github.com/ucloud/ucloud-sdk-go/ucloud/version"
)

const DefaultDasURL = "https://das-rpt.ucloud.cn/log"

//Logger 日志
var logger *log.Logger
var mu sync.Mutex
var out = Cxt.GetWriter()
var tracer = Tracer{DefaultDasURL}

func initConfigDir() {
	if _, err := os.Stat(GetLogFileDir()); os.IsNotExist(err) {
		err := os.MkdirAll(GetLogFileDir(), LocalFileMode)
		if err != nil {
			panic(err)
		}
	}
}

func initLog() error {
	initConfigDir()
	file, err := os.OpenFile(GetLogFilePath(), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("open log file failed: %v", err)
	}
	logger = log.New()
	logger.SetNoLock()
	logger.AddHook(NewLogRotateHook(file))
	logger.SetOutput(file)

	return nil
}

func logCmd() {
	args := make([]string, len(os.Args))
	copy(args, os.Args)
	for idx, arg := range args {
		for _, word := range []string{"password", "private-key", "public-key"} {
			if strings.Contains(arg, word) && idx <= len(args)-2 {
				args[idx+1] = strings.Repeat("*", 8)
			}
		}
	}
	LogInfo(fmt.Sprintf("command: %s", strings.Join(args, " ")))
}

//GetLogger return point of logger
func GetLogger() *log.Logger {
	return logger
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
	_, ok := os.LookupEnv("COMP_LINE")
	if ok {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	goID := curGoroutineID()
	for _, line := range logs {
		logger.WithField("goroutine_id", goID).Info(line)
	}
	if ConfigIns.AgreeUploadLog {
		UploadLogs(logs, "info", goID)
	}
}

//LogPrint 记录日志
func LogPrint(logs ...string) {
	_, ok := os.LookupEnv("COMP_LINE")
	if ok {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	goID := curGoroutineID()
	for _, line := range logs {
		logger.WithField("goroutine_id", goID).Print(line)
		fmt.Fprintln(out, line)
	}
	if ConfigIns.AgreeUploadLog {
		UploadLogs(logs, "print", goID)
	}
}

//LogWarn 记录日志
func LogWarn(logs ...string) {
	_, ok := os.LookupEnv("COMP_LINE")
	if ok {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	goID := curGoroutineID()
	for _, line := range logs {
		logger.WithField("goroutine_id", goID).Warn(line)
		fmt.Fprintln(out, line)
	}
	if ConfigIns.AgreeUploadLog {
		UploadLogs(logs, "warn", goID)
	}
}

//LogError 记录日志
func LogError(logs ...string) {
	_, ok := os.LookupEnv("COMP_LINE")
	if ok {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	goID := curGoroutineID()
	for _, line := range logs {
		logger.WithField("goroutine_id", goID).Error(line)
		fmt.Fprintln(out, line)
	}
	if ConfigIns.AgreeUploadLog {
		UploadLogs(logs, "error", goID)
	}
}

//UploadLogs send logs to das server
func UploadLogs(logs []string, level string, goID int64) {
	var lines []string
	for _, log := range logs {
		line := fmt.Sprintf("time=%s level=%s goroutine_id=%d msg=%s", time.Now().Format(time.RFC3339Nano), level, goID, log)
		lines = append(lines, line)
	}
	tracer.Send(lines)
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

//Tracer upload log to server if allowed
type Tracer struct {
	DasUrl string
}

func (t Tracer) wrapLogs(log []string) ([]byte, error) {
	dataSet := make([]map[string]interface{}, 0)
	dataItem := map[string]interface{}{
		"level": "info",
		"topic": "api",
		"log":   log,
	}
	dataSet = append(dataSet, dataItem)
	reqUUID := uuid.NewV4()
	sessionID := uuid.NewV4()
	user, err := GetUserInfo()
	if err != nil {
		return nil, err
	}
	payload := map[string]interface{}{
		"aid":  "iywtleaa",
		"uuid": reqUUID,
		"sid":  sessionID,
		"ds":   dataSet,
		"cs": map[string]interface{}{
			"uname": user.UserEmail,
		},
	}
	marshaled, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("cannot to marshal log: %s", err)
	}
	return marshaled, nil
}

//Send logs to server
func (t Tracer) Send(logs []string) error {
	body, err := t.wrapLogs(logs)
	if err != nil {
		return err
	}
	for i := 0; i < len(body); i++ {
		body[i] = ^body[i]
	}

	client := &http.Client{}
	ua := fmt.Sprintf("GO/%s GO-SDK/%s %s", runtime.Version(), version.Version, UserAgent)
	req, err := http.NewRequest("POST", t.DasUrl, bytes.NewReader(body))
	req.Header.Add("Origin", "https://sdk.ucloud.cn")
	req.Header.Add("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("send logs failed: status %d %s", resp.StatusCode, resp.Status)
	}

	return nil
}
