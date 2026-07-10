package uk8s

import (
	"encoding/base64"
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

// requiredNodeAddArgs returns the minimum flag set required for the
// uk8s node add command to pass cobra's MarkFlagRequired checks and
// PreRunE validation. IDs use the "id/name" form so PickResourceID
// is exercised (matching the create_test contract).
func requiredNodeAddArgs() []string {
	return []string{
		"add",
		"--cluster-id", "uk8s-a/name",
		"--cpu", "2",
		"--memory-mb", "4096",
		"--count", "1",
		"--charge-type", "Dynamic",
		"--password", "Password1",
	}
}

// TestNodeAddRequestMatchesDocument mirrors TestCreateRequestMatchesDocument:
// with all optional flags set, the SDK request must match the documented
// field names, base64-encoded password/userdata/initscript, and stripped
// "id/name" suffixes.
func TestNodeAddRequestMatchesDocument(t *testing.T) {
	ctx, requests := setupCommandGateway(t)
	cmd := newNode(ctx)
	args := append(requiredNodeAddArgs(),
		"--machine-type", "G",
		"--gpu", "1",
		"--gpu-type", "V100",
		"--image-id", "uimage-a/name",
		"--subnet-id", "subnet-a/name",
		"--nodegroup-id", "uk8sng-a/name",
		"--isolation-group-id", "ig-a/name",
		"--group", "Default",
		"--max-pods", "110",
		"--labels", "env=test,team=cli",
		"--taints", "dedicated=test:NoSchedule",
		"--user-data", "cloud-init",
		"--init-script", "echo ready",
		"--cpu-platform", "Intel/Cascadelake",
		"--boot-disk-type", "CLOUD_SSD",
		"--boot-disk-size-gb", "40",
		"--data-disk-type", "CLOUD_SSD",
		"--data-disk-size-gb", "100",
	)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute node add: %v", err)
	}

	want := map[string]string{
		"Action":             "AddUK8SUHostNode",
		"Region":             "cn-sh2",
		"ProjectId":          "org-test",
		"ClusterId":          "uk8s-a",
		"CPU":                "2",
		"Mem":                "4096",
		"Count":              "1",
		"ChargeType":         "Dynamic",
		"Password":           base64.StdEncoding.EncodeToString([]byte("Password1")),
		"MachineType":        "G",
		"GPU":                "1",
		"GpuType":            "V100",
		"ImageId":            "uimage-a",
		"SubnetId":           "subnet-a",
		"NodeGroupId":        "uk8sng-a",
		"IsolationGroup":     "ig-a",
		"Tag":                "Default",
		"MaxPods":            "110",
		"Labels":             "env=test,team=cli",
		"Taints":             "dedicated=test:NoSchedule",
		"UserData":           base64.StdEncoding.EncodeToString([]byte("cloud-init")),
		"InitScript":         base64.StdEncoding.EncodeToString([]byte("echo ready")),
		"MinimalCpuPlatform": "Intel/Cascadelake",
		"BootDiskType":       "CLOUD_SSD",
		"BootDiskSize":       "40",
		"DataDiskType":       "CLOUD_SSD",
		"DataDiskSize":       "100",
	}
	got := lastRequest(t, requests)
	assertRequest(t, got, want)
}

// TestNodeAddOmitsDocumentedOptionalFields mirrors
// TestCreateOmitsDocumentedOptionalFields: with only required flags,
// optional string fields must be absent from the request form so the SDK
// does not send empty-string values that confuse the backend. Int fields
// (GPU, MaxPods, BootDiskSize, DataDiskSize, Quantity) and the DisableSchedule
// bool all default to 0/false and the SDK marshals them as such — create
// avoids this with an explicit "optional" map that nils unset values, but
// node_add does not. Asserting their omission here would force the same
// pattern; left as a follow-up since it would change the SDK wire format
// of an existing command.
func TestNodeAddOmitsDocumentedOptionalFields(t *testing.T) {
	ctx, requests := setupCommandGateway(t)
	cmd := newNode(ctx)
	cmd.SetArgs(requiredNodeAddArgs())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute node add: %v", err)
	}
	got := lastRequest(t, requests)

	for _, key := range []string{
		"MachineType", "GpuType", "ImageId", "SubnetId", "NodeGroupId",
		"IsolationGroup", "Tag", "Labels", "Taints", "UserData",
		"InitScript", "MinimalCpuPlatform", "BootDiskType",
		"DataDiskType",
	} {
		if _, ok := got[key]; ok {
			t.Errorf("optional request field %s must be omitted", key)
		}
	}
}

// TestNodeAddRejectsInvalidShapeBeforeRequest mirrors
// TestCreateRejectsInvalidShapeBeforeRequest: mutate one flag at a time
// and assert PreRunE rejects the mutation before any API call lands.
// Mutators operate through *[]string so they can append flags absent from
// requiredNodeAddArgs (e.g. machine-type, boot-disk-size-gb).
func TestNodeAddRejectsInvalidShapeBeforeRequest(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*[]string)
		wantErr string
	}{
		{
			name: "cpu below range",
			mutate: func(a *[]string) { setArg(a, "--cpu", "1") },
			wantErr: "--cpu must be between 2 and 64",
		},
		{
			name: "memory not multiple of 1024",
			mutate: func(a *[]string) { setArg(a, "--memory-mb", "5000") },
			wantErr: "multiple of 1024",
		},
		{
			name: "count above range",
			mutate: func(a *[]string) { setArg(a, "--count", "100") },
			wantErr: "--count must be between 1 and 50",
		},
		{
			name: "charge-type unknown",
			mutate: func(a *[]string) { setArg(a, "--charge-type", "PayAsYouGo") },
			wantErr: "--charge-type must be one of",
		},
		{
			name: "machine-type G without gpu",
			mutate: func(a *[]string) { setArg(a, "--machine-type", "G") },
			wantErr: "--gpu and --gpu-type are required",
		},
		{
			name: "boot disk size below range",
			mutate: func(a *[]string) { setArg(a, "--boot-disk-size-gb", "10") },
			wantErr: "--boot-disk-size-gb must be between 40 and 500",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, requests := setupCommandGateway(t)
			cmd := newNode(ctx)
			args := requiredNodeAddArgs()
			tc.mutate(&args)
			cmd.SetArgs(args)
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
			}
			if len(*requests) != 0 {
				t.Fatal("invalid mutation must not reach the API")
			}
		})
	}
}

// setArg replaces the value of --flag in args if present, or appends
// --flag value if absent. Operates on *[]string so appends (which may
// reallocate the backing array) propagate to the caller.
func setArg(argsPtr *[]string, flag, value string) {
	args := *argsPtr
	for i := range args {
		if args[i] == flag {
			args[i+1] = value
			return
		}
	}
	*argsPtr = append(args, flag, value)
}
