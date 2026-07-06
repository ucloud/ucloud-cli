package cli_test

import (
	"errors"
	"testing"

	"github.com/ucloud/ucloud-cli/pkg/cli"
)

func TestAllRegionsForwardsProviderAndError(t *testing.T) {
	ctx := cli.NewContext(cli.Deps{
		AllRegions: func() ([]string, error) { return []string{"cn-bj2", "us-ca"}, nil },
	})
	regions, err := ctx.AllRegions()
	if err != nil || len(regions) != 2 {
		t.Fatalf("AllRegions = %v, %v; want 2 regions, nil err", regions, err)
	}

	// error must propagate (so --all-region reports a region-fetch failure
	// instead of silently listing nothing).
	wantErr := errors.New("fetch region failed")
	ctxErr := cli.NewContext(cli.Deps{
		AllRegions: func() ([]string, error) { return nil, wantErr },
	})
	if _, err := ctxErr.AllRegions(); !errors.Is(err, wantErr) {
		t.Fatalf("AllRegions error = %v, want %v", err, wantErr)
	}

	// nil-safe when no provider injected.
	empty := cli.NewContext(cli.Deps{})
	if r, err := empty.AllRegions(); r != nil || err != nil {
		t.Fatalf("nil provider: want (nil,nil), got (%v,%v)", r, err)
	}
}
