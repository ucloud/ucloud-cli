package cli_test

import (
	"strings"
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func TestContextLogForwarders(t *testing.T) {
	// base.Log* early-returns under COMP_LINE without the initialized global
	// logger, so the forwarders are exercised without a panic.
	t.Setenv("COMP_LINE", "1")
	ctx := cli.NewContext(cli.Deps{})

	// Non-request product diagnostics: must not panic.
	ctx.LogInfo("info")
	ctx.LogPrint("print")
	ctx.LogWarn("warn")
	ctx.LogError("err")

	if !strings.Contains(ctx.LogFilePath(), "cli.log") {
		t.Fatalf("LogFilePath = %q, want it to contain cli.log", ctx.LogFilePath())
	}
}
