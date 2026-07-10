//go:build live
// +build live

package cmd

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	svcuhost "github.com/ucloud/ucloud-sdk-go/services/uhost"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/cmd/internal/platform"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
	"github.com/ucloud/ucloud-cli/products/uhost"
)

// uhost_test.go drives the live UHost flow through the migrated products/uhost
// command tree (uhost moved out of cmd in Part 6). It hits the real API,
// creates paid resources, and needs valid credentials, so it is gated behind
// the `live` build tag. Run it explicitly with:
// `go test -tags live ./cmd -run '^TestUhost$' -count=1`.
// The image-id lookup the old test did via
// the cmd-local NewCmdUImageList/ImageRow shim is now a direct DescribeImage SDK
// call (image is served by the uhost SDK). create/delete narration now flows
// through ctx.NewProgress → ctx.ProgressWriter (the ctx Out buffer in table
// mode) instead of the old global progress document, so the test captures the
// ctx Out buffer.

// fetchLiveImageID returns the first Available Base image id via DescribeImage.
func fetchLiveImageID(t *testing.T) string {
	client := newServiceClient(svcuhost.NewClient)
	req := client.NewDescribeImageRequest()
	req.ImageType = sdk.String("Base")
	resp, err := client.DescribeImage(req)
	if err != nil {
		t.Fatalf("unexpected error fetching image list: %v", err)
	}
	for _, image := range resp.ImageSet {
		if image.State == "Available" {
			return image.ImageId
		}
	}
	t.Fatalf("image list is empty")
	return ""
}

func TestUhost(t *testing.T) {
	platform.InitConfig()
	var out bytes.Buffer
	// Buffer-backed ctx (table mode): create/delete narration via ctx.NewProgress
	// routes to ProgressWriter == Out; the cmd-package completion providers + real
	// config preserve the live behaviour.
	ctx := cli.NewContext(cli.Deps{
		In:     strings.NewReader(""),
		Out:    &out,
		Err:    &out,
		Format: cli.OutputTable,
		DefaultsProvider: func() command.Defaults {
			return command.Defaults{Region: platform.ConfigIns.Region, Zone: platform.ConfigIns.Zone, ProjectID: platform.ConfigIns.ProjectID}
		},
		RegionList:      getRegionList,
		ZoneList:        getZoneList,
		ProjectList:     getProjectList,
		AllRegions:      getAllRegions,
		ClientConfig:    func() *sdk.Config { return platform.ClientConfig },
		BuildCredential: platform.BuildCredential,
		AttachHandlers:  platform.AttachHandlers,
	})
	root := topLevelCmd(t, uhost.New().NewCommand(ctx), "uhost")

	imageID := fetchLiveImageID(t)

	runE := func(name string, flags []string) (string, error) {
		out.Reset()
		subCmd(t, root, name)
		root.SetArgs(append([]string{name}, flags...))
		if err := root.Execute(); err != nil {
			return out.String(), fmt.Errorf("unexpected error executing %s: %w, flags: %v", name, err, flags)
		}
		return out.String(), nil
	}
	run := func(name string, flags []string) string {
		content, err := runE(name, flags)
		if err != nil {
			t.Fatalf("%v, output: %s", err, content)
		}
		return content
	}

	createOut := run("create", []string{
		"--zone=cn-bj2-03",
		"--cpu=1",
		"--memory-gb=1",
		"--image-id=" + imageID,
		"--password=testlxj@123",
		"--hot-plug=false",
		"--data-disk-type=NONE",
	})
	createRe := regexp.MustCompile(`uhost\[([\w-]+)\] is initializing\.\.\.done`)
	m := createRe.FindStringSubmatch(createOut)
	if m == nil {
		t.Errorf("unexpect create output:%s", createOut)
		return
	}
	uhostID := m[1]
	idFlag := fmt.Sprintf("--uhost-id=%s", uhostID)
	deleted := false
	defer func() {
		if deleted {
			return
		}
		content, err := runE("delete", []string{"--yes", "--destroy", idFlag})
		if err != nil {
			t.Logf("cleanup delete failed for %s: %v, output: %s", uhostID, err, content)
			return
		}
		t.Logf("cleanup delete succeeded for %s: %s", uhostID, content)
	}()

	assertRun := func(name string, flags []string, re *regexp.Regexp) {
		content := run(name, flags)
		if re.FindStringSubmatch(content) == nil {
			t.Errorf("unexpect %s output:%s", name, content)
		}
	}

	assertRun("restart", []string{idFlag}, regexp.MustCompile(`uhost\[([\w-]+)\] is restarting\.\.\.done`))
	assertRun("poweroff", []string{"--yes", idFlag}, regexp.MustCompile(`uhost\[([\w-]+)\] is power off`))

	time.Sleep(time.Second * 5)
	assertRun("start", []string{idFlag}, regexp.MustCompile(`uhost\[([\w-]+)\] is starting\.\.\.done`))
	assertRun("stop", []string{idFlag}, regexp.MustCompile(`uhost\[([\w-]+)\] is shutting down\.\.\.done`))
	run("delete", []string{"--yes", "--destroy", idFlag})
	deleted = true
}
