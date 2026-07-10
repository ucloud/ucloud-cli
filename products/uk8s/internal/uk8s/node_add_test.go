package uk8s

import (
	"strings"
	"testing"
)

// TestNodeAddRejectsBadPasswordBeforeRequest guards against reverting
// node_add.go's PreRunE back to a single-class uppercase check. The rule
// must match uk8s_create's validatePassword (8-30 chars, allowed set,
// at least 2 of {uppercase, lowercase, digit, special}).
func TestNodeAddRejectsBadPasswordBeforeRequest(t *testing.T) {
	cases := []struct {
		name    string
		pwd     string
		wantErr string
	}{
		{name: "too short", pwd: "Ab1!", wantErr: "8-30"},
		{name: "illegal char", pwd: "Password 1", wantErr: "illegal characters"},
		{name: "single class", pwd: "abcdefgh", wantErr: "at least 2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, requests := setupCommandGateway(t)
			cmd := newNode(ctx)
			cmd.SetArgs([]string{
				"add",
				"--cluster-id", "uk8s-a",
				"--cpu", "2",
				"--memory-mb", "4096",
				"--count", "1",
				"--charge-type", "Dynamic",
				"--password", tc.pwd,
			})
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
			if len(*requests) != 0 {
				t.Fatal("invalid password must not reach the API")
			}
		})
	}
}
