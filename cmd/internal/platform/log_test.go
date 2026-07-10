// base/log_test.go
package platform

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRedactCmdArgs(t *testing.T) {
	args := []string{"ucloud", "login", "--private-key", "PRIKEY", "--code", "CODE1", "--token", "TOK1", "--authorization", "AUTH1"}
	got := strings.Join(redactCmdArgs(args), " ")
	for _, secret := range []string{"PRIKEY", "CODE1", "TOK1", "AUTH1"} {
		if strings.Contains(got, secret) {
			t.Errorf("redactCmdArgs leaked %q: %s", secret, got)
		}
	}
	if !strings.Contains(got, "login") {
		t.Error("non-sensitive args must be preserved")
	}
}

// 整行兜底：args 中内嵌的 URL query 形态的 code/token 也要被遮蔽
func TestRedactCmdArgsURLForm(t *testing.T) {
	args := []string{"ucloud", "x", "http://localhost/authorization?code=SEC&state=ST"}
	got := strings.Join(redactCmdArgs(args), " ")
	if strings.Contains(got, "SEC") {
		t.Errorf("url-embedded code leaked: %s", got)
	}
}

// 出口接线测试：直接走 LogInfo，确认脱敏真的接在出口上（防止 redactLogLines 调用行被误删而无测试失败）
func TestLogInfoOutletWired(t *testing.T) {
	if logger == nil {
		t.Fatal("logger not initialized by package init")
	}
	// COMP_LINE 存在时 Log* 直接 return，须确保未设置
	if v, ok := os.LookupEnv("COMP_LINE"); ok {
		os.Unsetenv("COMP_LINE")
		t.Cleanup(func() { os.Setenv("COMP_LINE", v) })
	}
	// 关闭上传路径，避免测试触网
	prevUpload := ConfigIns.AgreeUploadLog
	ConfigIns.AgreeUploadLog = false
	t.Cleanup(func() { ConfigIns.AgreeUploadLog = prevUpload })

	var buf bytes.Buffer
	prevOut := logger.Out
	logger.SetOutput(&buf)
	t.Cleanup(func() { logger.SetOutput(prevOut) })

	LogInfo(`Authorization: Bearer SECRET-WIRE`)

	got := buf.String()
	if got == "" {
		t.Fatal("LogInfo wrote nothing to logger output")
	}
	if strings.Contains(got, "SECRET-WIRE") {
		t.Errorf("LogInfo outlet leaked token: %s", got)
	}
	if !strings.Contains(got, "********") {
		t.Errorf("LogInfo outlet missing redaction placeholder: %s", got)
	}
}

// 扩面：任何经 Log* 出口的行都不得泄漏 token（HandleError → LogError 同样被覆盖）
func TestLogOutputsRedacted(t *testing.T) {
	lines := redactLogLines([]string{
		`request failed: Authorization: Bearer SECRET-AT`,
		`refresh response: {"access_token":"SECRET-AT2","refresh_token":"SECRET-RT"}`,
	})
	joined := strings.Join(lines, "\n")
	for _, s := range []string{"SECRET-AT", "SECRET-AT2", "SECRET-RT"} {
		if strings.Contains(joined, s) {
			t.Errorf("log line leaked %q: %s", s, joined)
		}
	}
}
