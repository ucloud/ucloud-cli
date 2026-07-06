package cmd

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/products/uhost"
)

// uhost_test.go drives the live UHost flow through the migrated products/uhost
// command tree (uhost moved out of cmd in Part 6). It hits the real API and
// needs valid credentials, so it is gated by name — run the rest of the suite
// with `go test ./... -skip TestUhost`. The image-id lookup the old test did via
// the cmd-local NewCmdUImageList/ImageRow shim is now a direct DescribeImage SDK
// call (image is served by the uhost SDK). create/delete narration now flows
// through ctx.NewProgress → ctx.ProgressWriter (the ctx Out buffer in table
// mode) instead of the global ux.Doc, so the test captures the ctx Out buffer.

// subCmd returns the child of root whose Use matches name.
func subCmd(t *testing.T, root *cobra.Command, name string) *cobra.Command {
	for _, c := range root.Commands() {
		if c.Use == name {
			return c
		}
	}
	t.Fatalf("uhost subcommand %q not found", name)
	return nil
}

// fetchLiveImageID returns the first Available Base image id via DescribeImage.
func fetchLiveImageID(t *testing.T) string {
	req := base.BizClient.NewDescribeImageRequest()
	req.ImageType = sdk.String("Base")
	resp, err := base.BizClient.DescribeImage(req)
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
	base.InitConfig()
	var out bytes.Buffer
	// Buffer-backed ctx (table mode): create/delete narration via ctx.NewProgress
	// routes to ProgressWriter == Out; the cmd-package completion providers + real
	// config preserve the live behaviour.
	ctx := cli.NewContext(cli.Deps{
		In:          strings.NewReader(""),
		Out:         &out,
		Err:         &out,
		Format:      cli.OutputTable,
		Config:      base.ConfigIns,
		RegionList:  getRegionList,
		ZoneList:    getZoneList,
		ProjectList: getProjectList,
		AllRegions:  getAllRegions,
	})
	root := uhost.New().NewCommand(ctx)

	imageID := fetchLiveImageID(t)

	run := func(name string, flags []string) string {
		out.Reset()
		c := subCmd(t, root, name)
		c.Flags().Parse(flags)
		if err := c.Execute(); err != nil {
			t.Fatalf("unexpected error executing %s: %v, flags: %v", name, err, flags)
		}
		return out.String()
	}

	createOut := run("create", []string{
		"--cpu=1",
		"--memory-gb=1",
		"--image-id=" + imageID,
		"--password=testlxj@123",
		"--hot-plug=false",
		"--create-eip-bandwidth-mb=10",
	})
	createRe := regexp.MustCompile(`uhost\[([\w-]+)\] which attached a data disk and binded an eip is initializing\.\.\.done`)
	m := createRe.FindStringSubmatch(createOut)
	if m == nil {
		t.Errorf("unexpect create output:%s", createOut)
		return
	}
	uhostID := m[1]
	idFlag := fmt.Sprintf("--uhost-id=%s", uhostID)

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
	assertRun("delete", []string{"--yes", idFlag}, regexp.MustCompile(`uhost\[([\w-]+)\] deleted`))
}
