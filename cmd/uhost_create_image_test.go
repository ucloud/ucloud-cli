package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	productuhost "github.com/ucloud/ucloud-cli/products/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
)

func saveBaseGlobalsForCreateImage(t *testing.T) {
	t.Helper()
	oldClientConfig, oldAuthCredential, oldConfigIns := base.ClientConfig, base.AuthCredential, base.ConfigIns
	t.Cleanup(func() {
		base.ClientConfig = oldClientConfig
		base.AuthCredential = oldAuthCredential
		base.ConfigIns = oldConfigIns
	})
}

func TestUhostCreateImageJSONEmitsStructuredResult(t *testing.T) {
	saveBaseGlobalsForCreateImage(t)

	const imageID = "uimage-json-contract"
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse request form: %v", err)
		}
		if got := r.Form.Get("Action"); got != "CreateCustomImage" {
			t.Fatalf("Action = %q, want CreateCustomImage", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"RetCode":0,"Action":"CreateCustomImageResponse","ImageId":%q}`, imageID)
	}))
	defer api.Close()

	cfg := sdk.NewConfig()
	cfg.BaseUrl = api.URL
	cfg.Region = "cn-bj2"
	cfg.Zone = "cn-bj2-03"
	cfg.ProjectId = "org-test"
	base.ClientConfig = &cfg
	base.AuthCredential = &base.CredentialConfig{PublicKey: "public", PrivateKey: "private"}
	base.ConfigIns = &base.AggConfig{ProjectID: "org-test", Region: "cn-bj2", Zone: "cn-bj2-03"}

	var stdout, stderr bytes.Buffer
	ctx := cli.NewContext(cli.Deps{
		Out:    &stdout,
		Err:    &stderr,
		Format: cli.OutputJSON,
		DefaultsProvider: func() command.Defaults {
			return command.Defaults{ProjectID: base.ConfigIns.ProjectID, Region: base.ConfigIns.Region, Zone: base.ConfigIns.Zone}
		},
		ClientConfig:    func() *sdk.Config { return base.ClientConfig },
		BuildCredential: base.BuildCredential,
		AttachHandlers:  base.AttachHandlers,
	})
	root := topLevelCmd(t, productuhost.New().NewCommand(ctx), "uhost")
	root.SetArgs([]string{
		"create-image",
		"--uhost-id", "uhost-for-image",
		"--image-name", "contract-image",
		"--async",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("create-image command failed: %v", err)
	}
	if !strings.Contains(stderr.String(), "iamge["+imageID+"] is making") {
		t.Fatalf("stderr progress = %q, want progress for %s", stderr.String(), imageID)
	}

	var rows []cli.OpResultRow
	if err := json.Unmarshal(stdout.Bytes(), &rows); err != nil {
		t.Fatalf("stdout must be JSON result rows, got %q: %v", stdout.String(), err)
	}
	want := []cli.OpResultRow{{ResourceID: imageID, Action: "create", Status: "Making"}}
	if len(rows) != 1 || rows[0] != want[0] {
		t.Fatalf("result rows = %#v, want %#v", rows, want)
	}
}
