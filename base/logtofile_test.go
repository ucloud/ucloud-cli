package base

import (
	"bytes"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

// logToFile writes to cli.log only — never to DAS telemetry — so the per-request
// logging handler does not upload every API request to the server.
func TestLogToFileWritesToLoggerOnly(t *testing.T) {
	// logToFile skips when COMP_LINE is present (completion); ensure it's absent.
	if v, ok := os.LookupEnv("COMP_LINE"); ok {
		os.Unsetenv("COMP_LINE")
		defer os.Setenv("COMP_LINE", v)
	}

	var buf bytes.Buffer
	old := logger
	logger = log.New()
	logger.SetOutput(&buf)
	defer func() { logger = old }()

	logToFile("api: DescribeUHostInstance, request: map[Region:cn-bj2]")

	if !strings.Contains(buf.String(), "DescribeUHostInstance") {
		t.Fatalf("logToFile did not write to the cli.log logger: %q", buf.String())
	}
}
