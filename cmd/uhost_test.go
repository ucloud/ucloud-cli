package cmd

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

type listUhostTest struct {
	expectedUhosts []string
	expectedOut    string
}

func (test listUhostTest) run(t *testing.T) {
	cmd := NewCmdUHostList()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command:%v", err)
	}
}

type createUHostTest struct {
	flags             []string
	uhostIds          []string
	expectedOutRegexp *regexp.Regexp
}

func (test createUHostTest) run(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	cmd := NewCmdUHostCreate(buf)
	cmd.Flags().Parse(test.flags)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command: %v, flags: %v", err, test.flags)
	}
	list := test.expectedOutRegexp.FindStringSubmatch(buf.String())
	if list == nil {
		t.Errorf("unexpect output:%s", buf.String())
	} else {
		if len(list) == 2 {
			test.uhostIds = append(test.uhostIds, list[1])
		}
	}
}

type deleteUHostTest struct {
	flags             []string
	uhostIds          []string
	expectedOutRegexp *regexp.Regexp
}

func (test deleteUHostTest) run(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	cmd := NewCmdUHostDelete(buf)
	cmd.Flags().Parse(test.flags)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command: %v, flags: %v", err, test.flags)
	}
	list := test.expectedOutRegexp.FindStringSubmatch(buf.String())
	if list == nil {
		t.Errorf("unexpect output:%s", buf.String())
	}
}

func TestCreateUhost(t *testing.T) {
	createT := createUHostTest{
		expectedOutRegexp: regexp.MustCompile(`uhost\[([\w-]+)\] is initializing...done`),
		flags: []string{
			"--cpu=1",
			"--memory-gb=1",
			"--image-id=uimage-aaee5e",
			"--password=test.lxj",
		},
	}
	createT.run(t)

	deleteT := deleteUHostTest{
		uhostIds:          createT.uhostIds,
		expectedOutRegexp: regexp.MustCompile(`uhost:\[[w-]+\] deleted`),
		flags:             []string{"--yes"},
	}
	deleteT.flags = append(deleteT.flags, fmt.Sprintf("--uhost-id=%s", strings.Join(deleteT.uhostIds, ",")))
	deleteT.run(t)
}
