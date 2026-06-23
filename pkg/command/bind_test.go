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

// flagCandidates returns the registered upstream completion candidates for a flag.
// cobra.Completion is an alias for string, so the result is a plain []string.
func flagCandidates(t *testing.T, cmd *cobra.Command, name string) []string {
	t.Helper()
	fn, ok := cmd.GetFlagCompletionFunc(name)
	if !ok || fn == nil {
		t.Fatalf("no completion registered for flag %q", name)
	}
	comps, _ := fn(cmd, nil, "")
	return comps
}

func TestSetCompletionRegisters(t *testing.T) {
	cmd := newCmd()
	command.SetCompletion(cmd, "f", func() []string { return []string{"a", "b"} })

	if got := flagCandidates(t, cmd, "f"); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("completion func returned %v, want [a b]", got)
	}
}

func TestSetFlagValuesRegisters(t *testing.T) {
	cmd := newCmd()
	command.SetFlagValues(cmd, "f", "x", "y")

	if got := flagCandidates(t, cmd, "f"); !reflect.DeepEqual(got, []string{"x", "y"}) {
		t.Fatalf("completion candidates = %v, want [x y]", got)
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
	// Completion registered (upstream).
	if _, ok := cmd.GetFlagCompletionFunc("region"); !ok {
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
	if got := flagCandidates(t, cmd, "charge-type"); !reflect.DeepEqual(got, []string{"Month", "Dynamic", "Year"}) {
		t.Fatalf("charge-type completion values = %v", got)
	}
}

// partialReq has only some of the optional reflection fields (no Limit/Offset).
type partialReq struct {
	request.CommonBase
	ChargeType *string
	Quantity   *int
}

func TestBindCommonParams(t *testing.T) {
	regionList := func() []string { return []string{"cn-bj2"} }
	zoneList := func(region string) []string { return []string{region} }
	projectList := func() []string { return []string{"org-x"} }
	def := command.Defaults{Region: "cn-bj2", Zone: "cn-bj2-02", ProjectID: "org-x"}

	cases := []struct {
		name    string
		req     interface{}
		want    []string // flags that MUST be registered
		notWant []string // flags that MUST NOT be registered
	}{
		{
			name:    "all fields present",
			req:     &fakeReq{},
			want:    []string{"region", "zone", "project-id", "limit", "offset", "charge-type", "quantity"},
			notWant: nil,
		},
		{
			name:    "missing limit and offset",
			req:     &partialReq{},
			want:    []string{"region", "zone", "project-id", "charge-type", "quantity"},
			notWant: []string{"limit", "offset"},
		},
		{
			name:    "only request.Common",
			req:     &request.CommonBase{},
			want:    []string{"region", "zone", "project-id"},
			notWant: []string{"limit", "offset", "charge-type", "quantity"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "x"}
			// Must NOT panic regardless of which optional fields the req carries.
			command.BindCommonParams(cmd, tc.req, def, regionList, zoneList, projectList)

			for _, name := range tc.want {
				if cmd.Flags().Lookup(name) == nil {
					t.Errorf("flag %q not registered, want registered", name)
				}
			}
			for _, name := range tc.notWant {
				if cmd.Flags().Lookup(name) != nil {
					t.Errorf("flag %q registered, want skipped", name)
				}
			}
		})
	}
}

func TestBindCommonParamsRefWiring(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	req := &fakeReq{}

	command.BindCommonParams(cmd, req,
		command.Defaults{Region: "cn-bj2"},
		func() []string { return []string{"cn-bj2"} },
		func(region string) []string { return []string{region} },
		func() []string { return []string{"org-x"} },
	)

	if err := cmd.Flags().Set("region", "cn-sh2"); err != nil {
		t.Fatalf("set region flag: %v", err)
	}
	if got := req.GetRegion(); got != "cn-sh2" {
		t.Fatalf("req.GetRegion() = %q, want cn-sh2 (ref wiring broken)", got)
	}
	if req.Limit == nil || *req.Limit != 100 {
		t.Fatalf("limit default not wired: %v", req.Limit)
	}
}
