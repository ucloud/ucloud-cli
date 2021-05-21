package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/ux"
)

type listUhostTest struct {
	expectedUhosts []string
	expectedOut    string
}

func (test listUhostTest) run(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := NewCmdUHostList(buf)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command:%v", err)
	}
}

type listImageTest struct {
	flags []string
}

func (test *listImageTest) run(t *testing.T) string {
	global.JSON = true
	buf := new(bytes.Buffer)
	cmd := NewCmdUImageList(buf)
	cmd.Flags().Parse(test.flags)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command: %v, flags: %v", err, test.flags)
	}

	var images []ImageRow
	err := json.Unmarshal(buf.Bytes(), &images)
	if err != nil {
		t.Fatalf("unexpected error of fetching image list: %v", err)
	}
	if len(images) == 0 {
		t.Fatalf("image list is empty")
	}
	// for _, image := range images {
	// 	// image.ImageName
	// }
	return images[0].ImageID
}

type createUHostTest struct {
	flags             []string
	uhostIDs          []string
	expectedOutRegexp *regexp.Regexp
}

func (test *createUHostTest) run(t *testing.T) {
	cmd := NewCmdUHostCreate()
	cmd.Flags().Parse(test.flags)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command: %v, flags: %v", err, test.flags)
	}
	lines := ux.Doc.Content()
	content := strings.Join(lines, "\n")
	list := test.expectedOutRegexp.FindStringSubmatch(content)
	if list == nil {
		t.Errorf("unexpect output:%s", content)
	} else {
		if len(list) == 2 {
			test.uhostIDs = append(test.uhostIDs, list[1])
		}
	}
}

type deleteUHostTest struct {
	flags             []string
	uhostIDs          []string
	expectedOutRegexp *regexp.Regexp
}

func (test *deleteUHostTest) run(t *testing.T) {
	cmd := NewCmdUHostDelete()
	cmd.Flags().Parse(test.flags)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command: %v, flags: %v", err, test.flags)
	}
	lines := ux.Doc.Content()
	content := strings.Join(lines, "\n")
	list := test.expectedOutRegexp.FindStringSubmatch(content)
	if list == nil {
		t.Errorf("unexpect output:%s", content)
	}
}

type stopUHostTest struct {
	flags             []string
	uhostIDs          []string
	expectedOutRegexp *regexp.Regexp
}

func (test *stopUHostTest) run(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := NewCmdUHostStop(buf)
	cmd.Flags().Parse(test.flags)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command: %v, flags: %v", err, test.flags)
	}
	list := test.expectedOutRegexp.FindStringSubmatch(buf.String())
	if list == nil {
		t.Errorf("unexpect output:%s", buf.String())
	}
}

type startUHostTest struct {
	flags             []string
	uhostIDs          []string
	expectedOutRegexp *regexp.Regexp
}

func (test *startUHostTest) run(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := NewCmdUHostStart(buf)
	cmd.Flags().Parse(test.flags)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command: %v, flags: %v", err, test.flags)
	}
	list := test.expectedOutRegexp.FindStringSubmatch(buf.String())
	if list == nil {
		t.Errorf("unexpect output:%s", buf.String())
	}
}

type restartUHostTest struct {
	flags             []string
	uhostIDs          []string
	expectedOutRegexp *regexp.Regexp
}

func (test *restartUHostTest) run(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := NewCmdUHostReboot(buf)
	cmd.Flags().Parse(test.flags)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command: %v, flags: %v", err, test.flags)
	}
	list := test.expectedOutRegexp.FindStringSubmatch(buf.String())
	if list == nil {
		t.Errorf("unexpect output:%s", buf.String())
	}
}

type poweroffUHostTest struct {
	flags             []string
	uhostIDs          []string
	expectedOutRegexp *regexp.Regexp
}

func (test *poweroffUHostTest) run(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd := NewCmdUHostPoweroff(buf)
	cmd.Flags().Parse(test.flags)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command: %v, flags: %v", err, test.flags)
	}
	list := test.expectedOutRegexp.FindStringSubmatch(buf.String())
	if list == nil {
		t.Errorf("unexpect output:%s", buf.String())
	}
}

func TestUhost(t *testing.T) {
	base.InitConfig()
	listImageT := listImageTest{
		flags: []string{"--json"},
	}
	imageID := listImageT.run(t)

	createT := createUHostTest{expectedOutRegexp: regexp.MustCompile(`uhost\[([\w-]+)\] which attached a data disk and binded an eip is initializing\.\.\.done`),
		flags: []string{
			"--cpu=1",
			"--memory-gb=1",
			"--image-id=" + imageID,
			"--password=testlxj@123",
			"--hot-plug=false",
			"--create-eip-bandwidth-mb=10",
		},
	}
	createT.run(t)

	restartT := restartUHostTest{
		flags:             []string{fmt.Sprintf("--uhost-id=%s", strings.Join(createT.uhostIDs, ","))},
		expectedOutRegexp: regexp.MustCompile(`uhost\[([\w-]+)\] is restarting\.\.\.done`),
	}
	restartT.run(t)

	poweroffT := poweroffUHostTest{
		flags:             []string{"--yes", fmt.Sprintf("--uhost-id=%s", strings.Join(createT.uhostIDs, ","))},
		expectedOutRegexp: regexp.MustCompile(`uhost\[([\w-]+)\] is power off`),
	}
	poweroffT.run(t)

	time.Sleep(time.Second * 5)
	startT := startUHostTest{
		flags:             []string{fmt.Sprintf("--uhost-id=%s", strings.Join(createT.uhostIDs, ","))},
		expectedOutRegexp: regexp.MustCompile(`uhost\[([\w-]+)\] is starting\.\.\.done`),
	}
	startT.run(t)

	stopT := stopUHostTest{
		flags:             []string{fmt.Sprintf("--uhost-id=%s", strings.Join(createT.uhostIDs, ","))},
		expectedOutRegexp: regexp.MustCompile(`uhost\[([\w-]+)\] is shutting down\.\.\.done`),
	}

	stopT.run(t)

	deleteT := deleteUHostTest{
		uhostIDs:          createT.uhostIDs,
		expectedOutRegexp: regexp.MustCompile(`uhost\[([\w-]+)\] deleted`),
		flags:             []string{"--yes"},
	}
	deleteT.flags = append(deleteT.flags, fmt.Sprintf("--uhost-id=%s", strings.Join(deleteT.uhostIDs, ",")))
	deleteT.run(t)

}
