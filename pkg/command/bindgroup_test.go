package command_test

import (
	"testing"

	"github.com/spf13/cobra"

	"github.com/ucloud/ucloud-cli/pkg/command"
)

func TestBindGroup(t *testing.T) {
	type req struct{ Tag *string }
	cmd := &cobra.Command{Use: "x"}
	r := &req{}

	command.BindGroup(cmd, r)

	if cmd.Flags().Lookup("group") == nil {
		t.Fatal("--group flag not registered")
	}
	if r.Tag == nil {
		t.Fatal("req.Tag not bound")
	}
	if *r.Tag != "" {
		t.Fatalf("req.Tag default = %q, want empty", *r.Tag)
	}
}
