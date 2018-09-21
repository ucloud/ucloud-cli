package cmd

import (
	"testing"
)

type listUhostTest struct {
	expectedUhosts []string
	expectedOut    string
}

func (test listUhostTest) run(t *testing.T) {
	cmd := NewCmdUHostList()
	cmd.SetArgs([]string{"--project-id", "org-4nfe1i"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing command:%v", err)
	}
}

func TestListUhost(t *testing.T) {
	test := listUhostTest{}
	test.run(t)
}
