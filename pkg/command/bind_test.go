package command_test

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-sdk-go/ucloud/request"

	"github.com/ucloud/ucloud-cli/pkg/command"
)

func newCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "x"}
	cmd.Flags().String("f", "", "")
	return cmd
}

func TestSetCompletionRegisters(t *testing.T) {
	cmd := newCmd()
	command.SetCompletion(cmd, "f", func() []string { return []string{"a", "b"} })

	fn := cmd.Flags().GetFlagValuesFunc("f")
	if fn == nil {
		t.Fatal("expected completion func to be registered, got nil")
	}
	if got := fn(); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("completion func returned %v, want [a b]", got)
	}
}

func TestSetFlagValuesRegisters(t *testing.T) {
	cmd := newCmd()
	command.SetFlagValues(cmd, "f", "x", "y")

	if got := cmd.Flags().GetFlagValues("f"); !reflect.DeepEqual(got, []string{"x", "y"}) {
		t.Fatalf("GetFlagValues returned %v, want [x y]", got)
	}
}

func TestBindRegionDefaultAndRef(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	req := &request.CommonBase{}

	command.BindRegion(cmd, req, command.Defaults{Region: "cn-bj2"}, func() []string { return []string{"cn-bj2"} })

	flag := cmd.Flags().Lookup("region")
	if flag == nil {
		t.Fatal("region flag not registered")
	}
	if flag.DefValue != "cn-bj2" {
		t.Fatalf("region default = %q, want cn-bj2", flag.DefValue)
	}
	// Completion registered.
	if cmd.Flags().GetFlagValuesFunc("region") == nil {
		t.Fatal("region completion func not registered")
	}
	// Ref wiring: setting the flag must update req's region (shared storage).
	if err := cmd.Flags().Set("region", "cn-sh2"); err != nil {
		t.Fatalf("set region flag: %v", err)
	}
	if got := req.GetRegion(); got != "cn-sh2" {
		t.Fatalf("req.GetRegion() = %q, want cn-sh2 (ref wiring broken)", got)
	}
}

func TestBindZoneEmptyDefault(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	req := &request.CommonBase{}

	command.BindZoneEmpty(cmd, req, func(region string) []string { return []string{region} })

	flag := cmd.Flags().Lookup("zone")
	if flag == nil {
		t.Fatal("zone flag not registered")
	}
	if flag.DefValue != "" {
		t.Fatalf("zone default = %q, want empty", flag.DefValue)
	}
}

// fakeReq exercises the reflection-based binders (Limit/Offset/ChargeType/Quantity).
type fakeReq struct {
	request.CommonBase
	Limit      *int
	Offset     *int
	ChargeType *string
	Quantity   *int
}

func TestBindLimitOffsetChargeTypeQuantity(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	req := &fakeReq{}

	command.BindLimit(cmd, req)
	command.BindOffset(cmd, req)
	command.BindChargeType(cmd, req)
	command.BindQuantity(cmd, req)

	if req.Limit == nil || *req.Limit != 100 {
		t.Fatalf("limit default not wired: %v", req.Limit)
	}
	if req.Offset == nil || *req.Offset != 0 {
		t.Fatalf("offset default not wired: %v", req.Offset)
	}
	if req.ChargeType == nil || *req.ChargeType != "Month" {
		t.Fatalf("charge-type default not wired: %v", req.ChargeType)
	}
	if req.Quantity == nil || *req.Quantity != 1 {
		t.Fatalf("quantity default not wired: %v", req.Quantity)
	}
	if got := cmd.Flags().GetFlagValues("charge-type"); !reflect.DeepEqual(got, []string{"Month", "Dynamic", "Year"}) {
		t.Fatalf("charge-type completion values = %v", got)
	}
}
