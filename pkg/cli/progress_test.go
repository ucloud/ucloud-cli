package cli_test

import (
	"bytes"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func TestProgressRefreshWritesToCtxWriter(t *testing.T) {
	var out bytes.Buffer
	ctx := cli.NewContext(cli.Deps{Out: &out, Err: &out, Format: cli.OutputTable})

	p := ctx.NewProgress()
	if p == nil {
		t.Fatal("NewProgress returned nil")
	}
	p.Refresh("total:2, success:1, fail:0")

	if !strings.Contains(out.String(), "total:2") {
		t.Fatalf("Refresh did not write to ctx writer: %q", out.String())
	}
}

func TestProgressNewBlock(t *testing.T) {
	var out bytes.Buffer
	ctx := cli.NewContext(cli.Deps{Out: &out, Format: cli.OutputJSON})

	p := ctx.NewProgress()
	if b := p.NewBlock(); b == nil {
		t.Fatal("NewBlock returned nil")
	}
}

func TestConcurrentActionRunsAllReqs(t *testing.T) {
	t.Setenv("COMP_LINE", "1") // base.LogInfo becomes a no-op (uninitialized global logger otherwise panics)
	var out bytes.Buffer
	ctx := cli.NewContext(cli.Deps{Out: &out, Err: &out, Format: cli.OutputJSON})

	var n int32
	actionFunc := func(req request.Common) (bool, []string) {
		atomic.AddInt32(&n, 1)
		return true, nil
	}
	reqs := []request.Common{&request.CommonBase{}, &request.CommonBase{}, &request.CommonBase{}}

	ctx.ConcurrentAction(reqs, 2, actionFunc)

	if got := atomic.LoadInt32(&n); got != 3 {
		t.Fatalf("ConcurrentAction ran actionFunc %d times, want 3", got)
	}
}
