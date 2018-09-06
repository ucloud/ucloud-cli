package cmd

import "testing"

func TestCmdRoot(t *testing.T) {
	root := NewCmdRoot()
	root.Execute()
}
