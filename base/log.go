package base

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	"github.com/ucloud/ucloud-cli/internal/common"
)

const DefaultDasURL = "https://das-rpt.ucloud.cn/log"

// Logger 日志
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

// redactCmdArgs 脱敏命令行参数：flag 值遮蔽（名单含 oauth 敏感词）+ 整体过 Redact 兜底
func redactCmdArgs(osArgs []string) []string {
	args := make([]string, len(osArgs))
	copy(args, osArgs)
	for idx, arg := range args {
		for _, word := range []string{"password", "private-key", "public-key", "code", "token", "authorization"} {
			if strings.Contains(arg, word) && idx <= len(args)-2 {
				args[idx+1] = strings.Repeat("*", 8)
			}
		}
	}
	for idx := range args {
		args[idx] = Redact(args[idx])
	}
	return args
}

// redactLogLines 日志出口统一脱敏（Phase 3 扩面：错误包装/调试输出经 Log* 的部分）
func redactLogLines(logs []string) []string {
	out := make([]string, len(logs))
	for i, line := range logs {
		out[i] = Redact(line)
	}
	return out
}

func logCmd() {
	args := redactCmdArgs(os.Args)
	LogInfo(fmt.Sprintf("command: %s", strings.Join(args, " ")))
}

// GetLogger return point of logger
func GetLogger() *log.Logger {
	return logger
}

// GetLogFileDir 获取日志文件路径
func GetLogFileDir() string {
	return common.GetHomePath() + fmt.Sprintf("/%s", ConfigPath)
}

// GetLogFilePath 获取日志文件路径
func GetLogFilePath() string {
	return common.GetHomePath() + fmt.Sprintf("/%s/cli.log", ConfigPath)
}

// logToFile writes lines to the local cli.log only — NO DAS telemetry upload —
// with the same redaction and COMP_LINE skip as LogInfo. Used by the platform
// request-logging handler so logging every API request does not inflate
// telemetry traffic for users who opted into log upload (see batch-1 plan
// Part 0 Task 0.2, decision A).
func logToFile(logs ...string) {
	if _, ok := os.LookupEnv("COMP_LINE"); ok {
		return
	}
	logs = redactLogLines(logs)
	mu.Lock()
	defer mu.Unlock()
	goID := curGoroutineID()
	for _, line := range logs {
		logger.WithField("goroutine_id", goID).Info(line)
	}
}

// LogInfo 记录日志
func LogInfo(logs ...string) {
	_, ok := os.LookupEnv("COMP_LINE")
	if ok {
		return
	}
	logs = redactLogLines(logs)
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

// LogPrint 记录日志. Console copy → global stdout; product code should prefer the
// ctx wrappers (→ *To with stderr) so machine output on stdout stays clean.
func LogPrint(logs ...string) { LogPrintTo(out, logs...) }

// LogPrintTo is LogPrint with a caller-chosen console writer w (file + telemetry
// unchanged).
func LogPrintTo(w io.Writer, logs ...string) {
	if _, ok := os.LookupEnv("COMP_LINE"); ok {
		return
	}
	logs = redactLogLines(logs)
	mu.Lock()
	defer mu.Unlock()
	goID := curGoroutineID()
	for _, line := range logs {
		logger.WithField("goroutine_id", goID).Print(line)
		fmt.Fprintln(w, line)
	}
	if ConfigIns.AgreeUploadLog {
		UploadLogs(logs, "print", goID)
	}
}

// LogWarn 记录日志. Console copy → global stdout; product code should prefer the
// ctx wrappers (→ *To with stderr).
func LogWarn(logs ...string) { LogWarnTo(out, logs...) }

// LogWarnTo is LogWarn with a caller-chosen console writer w (file + telemetry
// unchanged).
func LogWarnTo(w io.Writer, logs ...string) {
	if _, ok := os.LookupEnv("COMP_LINE"); ok {
		return
	}
	logs = redactLogLines(logs)
	mu.Lock()
	defer mu.Unlock()
	goID := curGoroutineID()
	for _, line := range logs {
		logger.WithField("goroutine_id", goID).Warn(line)
		fmt.Fprintln(w, line)
	}
	if ConfigIns.AgreeUploadLog {
		UploadLogs(logs, "warn", goID)
	}
}

// LogError 记录日志. The console copy goes to the global writer (stdout); product
// code should prefer ctx.HandleError (→ LogErrorTo with stderr) so machine output
// on stdout stays clean.
func LogError(logs ...string) {
	LogErrorTo(out, logs...)
}

// LogErrorTo is LogError with a caller-chosen console writer w; file logging and
// telemetry are unchanged. Products route the console copy to stderr via
// ctx.HandleError so stdout carries only machine-readable results.
func LogErrorTo(w io.Writer, logs ...string) {
	if _, ok := os.LookupEnv("COMP_LINE"); ok {
		return
	}
	logs = redactLogLines(logs)
	mu.Lock()
	defer mu.Unlock()
	goID := curGoroutineID()
	for _, line := range logs {
		logger.WithField("goroutine_id", goID).Error(line)
		fmt.Fprintln(w, line)
	}
	if ConfigIns.AgreeUploadLog {
		UploadLogs(logs, "error", goID)
	}
}

// UploadLogs send logs to das server
func UploadLogs(logs []string, level string, goID int64) {
	var lines []string
	for _, log := range logs {
		line := fmt.Sprintf("time=%s level=%s goroutine_id=%d msg=%s", time.Now().Format(time.RFC3339Nano), level, goID, log)
		lines = append(lines, line)
	}
	tracer.Send(lines)
}

// LogRotateHook rotate log file
type LogRotateHook struct {
	MaxSize int64
	Cut     float32
	LogFile *os.File
	mux     sync.Mutex
}

// Levels fires hook
func (hook *LogRotateHook) Levels() []log.Level {
	return log.AllLevels
}

// Fire do someting when hook is triggered
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

// NewLogRotateHook create a LogRotateHook
func NewLogRotateHook(file *os.File) *LogRotateHook {
	return &LogRotateHook{
		MaxSize: 1024 * 1024, //1MB
		Cut:     0.2,
		LogFile: file,
	}
}

// ToQueryMap tranform request to map
func ToQueryMap(req request.Common) map[string]string {
	reqMap, err := request.ToQueryMap(req)
	if err != nil {
		return nil
	}
	delete(reqMap, "Password")
	return reqMap
}

// requestLogLine formats an API request for the platform request-logging
// handler: "api: <Action>, request: <query map>" (Password already redacted by
// ToQueryMap). This replaces the per-command hand-rolled request logging that
// products used to build with ToQueryMap — every request is now logged
// uniformly at the SDK handler layer (see batch-1 plan Part 0 Task 0.2).
func requestLogLine(req request.Common) string {
	return fmt.Sprintf("api: %s, request: %v", req.GetAction(), ToQueryMap(req))
}

// Tracer upload log to server if allowed
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

// Send logs to server
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
