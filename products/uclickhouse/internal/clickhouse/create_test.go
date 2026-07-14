package clickhouse

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func TestCreateRejectsMistypedFlagValueBeforeAPI(t *testing.T) {
	cmd, apiCalled, cleanup := newCreateTestCommand(t)
	defer cleanup()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"create",
		"--clickhouse-machine-type-id", "s1-x1",
		"--data-disk-type", "--data-disk-type", "CLOUD_RSSD",
		"--clickhouse-version", "24.8.14.39",
		"--admin-password", "4277813Aa",
		"--async",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected create to reject mistyped flag value")
	}
	if *apiCalled {
		t.Fatal("create should reject malformed arguments before calling API")
	}
	if !strings.Contains(err.Error(), `flag --data-disk-type requires a value`) || !strings.Contains(err.Error(), `got "--data-disk-type"`) {
		t.Fatalf("error = %q, want data-disk-type value error", err.Error())
	}
}

func TestCreateRejectsMissingZookeeperOptionsBeforeAPI(t *testing.T) {
	cmd, apiCalled, cleanup := newCreateTestCommand(t)
	defer cleanup()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"create",
		"--clickhouse-machine-type-id", "s1-x1",
		"--name", "cli-jjk",
		"--data-disk-type", "CLOUD_RSSD",
		"--clickhouse-version", "24.8.14.39",
		"--admin-password", "4277813Aa",
		"--async",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected create to reject missing zookeeper options")
	}
	if *apiCalled {
		t.Fatal("create should reject missing zookeeper options before calling API")
	}
	for _, want := range []string{"--zookeeper-machine-type-id", "--zookeeper-data-disk-type", "--zookeeper-data-disk-size-gb"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want mention %s", err.Error(), want)
		}
	}
}

func TestCreateRejectsOtherMissingConditionalOptionsBeforeAPI(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "sec group IDs",
			args: []string{"--zookeeper-ha=false", "--sec-group", "true"},
			want: "--sec-group-ids",
		},
		{
			name: "multi zone names",
			args: []string{"--zookeeper-ha=false", "--multi-zone", "true"},
			want: "--multi-zone-name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, apiCalled, cleanup := newCreateTestCommand(t)
			defer cleanup()
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			baseArgs := []string{
				"create",
				"--clickhouse-machine-type-id", "s1-x1",
				"--data-disk-type", "CLOUD_RSSD",
				"--clickhouse-version", "24.8.14.39",
				"--admin-password", "4277813Aa",
				"--async",
			}
			cmd.SetArgs(append(baseArgs, tt.args...))

			err := cmd.Execute()
			if err == nil {
				t.Fatal("expected create to reject missing conditional options")
			}
			if *apiCalled {
				t.Fatal("create should reject missing conditional options before calling API")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %q, want mention %s", err.Error(), tt.want)
			}
		})
	}
}

func TestCreateRejectsBareBooleanArgument(t *testing.T) {
	cmd, apiCalled, cleanup := newCreateTestCommand(t)
	defer cleanup()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"create",
		"--clickhouse-machine-type-id", "s1-x1",
		"--data-disk-type", "CLOUD_RSSD",
		"--clickhouse-version", "24.8.14.39",
		"--admin-password", "4277813Aa",
		"--zookeeper-ha", "false",
		"--async",
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected create to reject bare boolean argument")
	}
	if *apiCalled {
		t.Fatal("create should reject bare boolean argument before calling API")
	}
	if !strings.Contains(err.Error(), "boolean flags must use --flag=false") {
		t.Fatalf("error = %q, want boolean flag hint", err.Error())
	}
}

func newCreateTestCommand(t *testing.T) (*cobra.Command, *bool, func()) {
	t.Helper()
	apiCalled := false
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"RetCode":0,"Message":"success","Data":{"ClusterId":"uck-test"}}`)
	}))

	cfg := sdk.NewConfig()
	cfg.BaseUrl = api.URL
	cfg.Region = "cn-bj2"
	cfg.ProjectId = "org-test"
	cred := auth.NewCredential()
	cred.PublicKey = "public"
	cred.PrivateKey = "private"

	var out, errOut bytes.Buffer
	ctx := cli.NewContext(cli.Deps{
		Out:    &out,
		Err:    &errOut,
		Format: cli.OutputJSON,
		DefaultsProvider: func() command.Defaults {
			return command.Defaults{ProjectID: "org-test", Region: "cn-bj2"}
		},
		ClientConfig: func() *sdk.Config {
			return &cfg
		},
		BuildCredential: func() *auth.Credential {
			return &cred
		},
		AttachHandlers: func(sdk.ServiceClient) {},
	})
	return NewCommand(ctx), &apiCalled, api.Close
}
