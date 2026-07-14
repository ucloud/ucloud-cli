package cli_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func TestContextLogForwarders(t *testing.T) {
	var err bytes.Buffer
	var infoCalls, printCalls, warnCalls, errorCalls int
	ctx := cli.NewContext(cli.Deps{
		Err: &err,
		LogInfo: func(logs ...string) {
			infoCalls += len(logs)
		},
		LogPrint: func(w io.Writer, logs ...string) {
			printCalls += len(logs)
			fmt.Fprint(w, strings.Join(logs, "\n"))
		},
		LogWarn: func(w io.Writer, logs ...string) {
			warnCalls += len(logs)
			fmt.Fprint(w, strings.Join(logs, "\n"))
		},
		LogError: func(w io.Writer, logs ...string) {
			errorCalls += len(logs)
			fmt.Fprint(w, strings.Join(logs, "\n"))
		},
		LogFilePath: func() string { return "/tmp/cli.log" },
	})

	ctx.LogInfo("info")
	ctx.LogPrint("print")
	ctx.LogWarn("warn")
	ctx.LogError("err")

	if infoCalls != 1 || printCalls != 1 || warnCalls != 1 || errorCalls != 1 {
		t.Fatalf("log provider calls = %d/%d/%d/%d, want all 1", infoCalls, printCalls, warnCalls, errorCalls)
	}
	if got := err.String(); !strings.Contains(got, "print") || !strings.Contains(got, "warn") || !strings.Contains(got, "err") {
		t.Fatalf("stderr log output = %q, want print/warn/err", got)
	}
	if !strings.Contains(ctx.LogFilePath(), "cli.log") {
		t.Fatalf("LogFilePath = %q, want it to contain cli.log", ctx.LogFilePath())
	}
}
