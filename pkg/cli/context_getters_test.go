package cli_test

import (
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func TestRegionZoneProjectListGetters(t *testing.T) {
	ctx := cli.NewContext(cli.Deps{
		RegionList:  func() []string { return []string{"cn-bj2"} },
		ZoneList:    func(r string) []string { return []string{r + "-01"} },
		ProjectList: func() []string { return []string{"org-x"} },
	})

	if got := ctx.RegionList(); len(got) != 1 || got[0] != "cn-bj2" {
		t.Fatalf("RegionList = %v", got)
	}
	if got := ctx.ZoneList("cn-bj2"); len(got) != 1 || got[0] != "cn-bj2-01" {
		t.Fatalf("ZoneList = %v", got)
	}
	if got := ctx.ProjectList(); len(got) != 1 || got[0] != "org-x" {
		t.Fatalf("ProjectList = %v", got)
	}

	// nil-safe when providers absent (non-standard-flag getters must not panic).
	empty := cli.NewContext(cli.Deps{})
	if empty.RegionList() != nil || empty.ZoneList("x") != nil || empty.ProjectList() != nil {
		t.Fatal("getters must be nil-safe when providers absent")
	}
}
