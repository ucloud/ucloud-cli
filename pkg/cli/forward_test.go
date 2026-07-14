package cli_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func TestContextForwarders(t *testing.T) {
	var buf bytes.Buffer
	ctx := cli.NewContext(cli.Deps{
		Out:        &buf,
		Format:     cli.OutputJSON,
		RegionList: func() []string { return []string{"r"} },
	})

	ctx.PrintList([]struct{ Name string }{{Name: "a"}})
	if !strings.Contains(buf.String(), `"Name"`) {
		t.Fatalf("PrintList(JSON) output missing \"Name\": %q", buf.String())
	}

	if got := ctx.PickResourceID("udb-x/n"); got != "udb-x" {
		t.Fatalf("PickResourceID = %q, want udb-x", got)
	}
}

// ctxFakeReq carries every optional reflection-bound field.
type ctxFakeReq struct {
	request.CommonBase
	Limit      *int
	Offset     *int
	ChargeType *string
	Quantity   *int
}

func TestContextBindCommonParams(t *testing.T) {
	ctx := cli.NewContext(cli.Deps{
		DefaultsProvider: func() command.Defaults {
			return command.Defaults{Region: "cn-bj2", Zone: "cn-bj2-02", ProjectID: "org-x"}
		},
		RegionList:  func() []string { return []string{"cn-bj2"} },
		ZoneList:    func(region string) []string { return []string{region} },
		ProjectList: func() []string { return []string{"org-x"} },
	})

	// Full request: every common flag must be registered, ctx defaults applied.
	cmd := &cobra.Command{Use: "x"}
	req := &ctxFakeReq{}
	ctx.BindCommonParams(cmd, req)

	for _, name := range []string{"region", "zone", "project-id", "limit", "offset", "charge-type", "quantity"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("flag %q not registered", name)
		}
	}
	if f := cmd.Flags().Lookup("region"); f == nil || f.DefValue != "cn-bj2" {
		t.Fatalf("region default not taken from ctx config: %+v", f)
	}

	// Request satisfying only request.Common: list/charge flags skipped, no panic.
	cmdCommon := &cobra.Command{Use: "y"}
	ctx.BindCommonParams(cmdCommon, &request.CommonBase{})
	for _, name := range []string{"limit", "offset", "charge-type", "quantity"} {
		if cmdCommon.Flags().Lookup(name) != nil {
			t.Errorf("flag %q registered for plain CommonBase, want skipped", name)
		}
	}
}

func TestContextPollerToReturnsProductCompatiblePoller(t *testing.T) {
	ctx := cli.NewContext(cli.Deps{
		NewPoller: func(describe func(string, *request.CommonBase) (interface{}, error), out io.Writer, opts ...cli.PollerOption) cli.Poller {
			return cli.NewPoller(describe, out, opts...)
		},
	})

	ctx.PollerTo(io.Discard, func(string, *request.CommonBase) (interface{}, error) {
		return struct{ State string }{State: "RUNNING"}, nil
	}).Spoll("res-1", "creating", []string{"RUNNING"})
}
