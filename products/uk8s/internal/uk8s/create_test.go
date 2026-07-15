package uk8s

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"

	"github.com/ucloud/ucloud-cli/pkg/cli"
	"github.com/ucloud/ucloud-cli/pkg/command"
)

func setupCreateMock(t *testing.T) (*cli.Context, *url.Values, func()) {
	t.Helper()

	values := &url.Values{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("parse request form: %v", err)
		}
		*values = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"RetCode":   0,
			"Action":    "CreateUK8SClusterV2Response",
			"ClusterId": "uk8s-test",
		})
	}))

	var out, errOut bytes.Buffer
	ctx := cli.NewContext(cli.Deps{
		In:     strings.NewReader(""),
		Out:    &out,
		Err:    &errOut,
		Format: cli.OutputTable,
		DefaultsProvider: func() command.Defaults {
			return command.Defaults{Region: "cn-sh2", ProjectID: "org-test"}
		},
		ClientConfig: func() *sdk.Config {
			return &sdk.Config{Region: "cn-sh2", ProjectId: "org-test", BaseUrl: server.URL}
		},
		BuildCredential: func() *auth.Credential {
			return &auth.Credential{PublicKey: "public", PrivateKey: "private"}
		},
		AttachHandlers: func(sdk.ServiceClient) {},
	})
	cleanup := func() {
		server.Close()
	}
	return ctx, values, cleanup
}

func requiredCreateArgs() []string {
	return []string{
		"--name", "demo-uk8s",
		"--password", "Password1",
		"--vpc-id", "uvnet-test/vpc-name",
		"--subnet-id", "subnet-test/subnet-name",
		"--service-cidr", "172.17.0.0/16",
		"--master-cpu", "2",
		"--master-memory-mb", "4096",
		"--master-machine-type", "N",
		"--master-zone", "cn-sh2-01",
		"--node-cpu", "2",
		"--node-count", "3",
		"--node-memory-mb", "4096",
		"--node-machine-type", "N",
		"--node-zone", "cn-sh2-01",
		"--image-id", "uimage-test/image-name",
		"--k8s-version", "1.34.5",
		"--async",
	}
}

func TestCreateRequestMatchesDocument(t *testing.T) {
	ctx, form, cleanup := setupCreateMock(t)
	defer cleanup()

	cmd := newCreate(ctx)
	args := append(requiredCreateArgs(),
		"--charge-type", "Month",
		"--quantity", "1",
		"--node-isolation-group-id", "ig-test",
		"--node-machine-type", "G",
		"--node-gpu", "1",
		"--node-gpu-type", "V100",
		"--node-labels", "env=test,team=cli",
		"--node-taints", "dedicated=test:NoSchedule",
		"--node-max-pods", "110",
		"--group", "Default",
		"--user-data", "cloud-init",
		"--init-script", "echo ready",
	)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute create: %v", err)
	}

	want := map[string]string{
		"Action":                 "CreateUK8SClusterV2",
		"Region":                 "cn-sh2",
		"ProjectId":              "org-test",
		"ClusterName":            "demo-uk8s",
		"Password":               base64.StdEncoding.EncodeToString([]byte("Password1")),
		"VPCId":                  "uvnet-test",
		"SubnetId":               "subnet-test",
		"ServiceCIDR":            "172.17.0.0/16",
		"K8sVersion":             "1.34.5",
		"ImageId":                "uimage-test",
		"MasterCPU":              "2",
		"MasterMem":              "4096",
		"MasterMachineType":      "N",
		"Master.0.Zone":          "cn-sh2-01",
		"Master.1.Zone":          "cn-sh2-01",
		"Master.2.Zone":          "cn-sh2-01",
		"Nodes.0.Zone":           "cn-sh2-01",
		"Nodes.0.CPU":            "2",
		"Nodes.0.Count":          "3",
		"Nodes.0.Mem":            "4096",
		"Nodes.0.MachineType":    "G",
		"Nodes.0.GPU":            "1",
		"Nodes.0.GpuType":        "V100",
		"Nodes.0.IsolationGroup": "ig-test",
		"Nodes.0.Labels":         "env=test,team=cli",
		"Nodes.0.Taints":         "dedicated=test:NoSchedule",
		"Nodes.0.MaxPods":        "110",
		"ChargeType":             "Month",
		"Quantity":               "1",
		"Tag":                    "Default",
		"UserData":               base64.StdEncoding.EncodeToString([]byte("cloud-init")),
		"InitScript":             base64.StdEncoding.EncodeToString([]byte("echo ready")),
	}
	for key, expected := range want {
		if got := form.Get(key); got != expected {
			t.Errorf("request %s = %q, want %q", key, got, expected)
		}
	}
}

func TestCreateOmitsDocumentedOptionalFields(t *testing.T) {
	ctx, form, cleanup := setupCreateMock(t)
	defer cleanup()

	cmd := newCreate(ctx)
	cmd.SetArgs(requiredCreateArgs())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute create: %v", err)
	}

	for _, key := range []string{
		"ChargeType", "Quantity", "UserData", "InitScript",
		"Nodes.0.MaxPods", "Nodes.0.IsolationGroup", "Nodes.0.Labels", "Nodes.0.Taints",
	} {
		if _, ok := (*form)[key]; ok {
			t.Errorf("optional request field %s must be omitted", key)
		}
	}
}

func TestCreateRejectsInvalidShapeBeforeRequest(t *testing.T) {
	ctx, form, cleanup := setupCreateMock(t)
	defer cleanup()

	cmd := newCreate(ctx)
	args := requiredCreateArgs()
	for i := range args {
		if args[i] == "--node-memory-mb" {
			args[i+1] = "5000"
		}
	}
	cmd.SetArgs(args)
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "multiple of 1024") {
		t.Fatalf("expected memory validation error, got %v", err)
	}
	if len(*form) != 0 {
		t.Fatal("invalid command must not reach the API")
	}
}

func TestValidatePassword(t *testing.T) {
	cases := []struct {
		name    string
		pwd     string
		wantErr string // substring; "" means accept
	}{
		// Valid
		{name: "9 chars upper+lower+digit", pwd: "Password1", wantErr: ""},
		{name: "8 chars lower+digit (boundary)", pwd: "abc12345", wantErr: ""},
		{name: "30 chars all four classes (boundary)", pwd: strings.Repeat("Aa1!", 7) + "Aa", wantErr: ""},
		{name: "9 chars lower+digit+special no upper", pwd: "password1!", wantErr: ""},
		{name: "all four classes", pwd: "Abc123!@#", wantErr: ""},
		{name: "backslash is allowed", pwd: `Pa\ssw0rd`, wantErr: ""},

		// Length
		{name: "7 chars too short", pwd: "Abc1234", wantErr: "8-30"},
		{name: "31 chars too long", pwd: strings.Repeat("Aa1!", 7) + "Aa1", wantErr: "8-30"},
		{name: "empty", pwd: "", wantErr: "8-30"},

		// Illegal chars (not in allowed set)
		{name: "contains space", pwd: "Pass word1", wantErr: "illegal characters"},
		{name: "contains tab", pwd: "Pass\tword1", wantErr: "illegal characters"},
		{name: "contains chinese char", pwd: "Password密1", wantErr: "illegal characters"},
		{name: "contains double-quote", pwd: `Pass"word1`, wantErr: "illegal characters"},
		{name: "contains backtick", pwd: "Pass`word1", wantErr: "illegal characters"},

		// Single class is not enough
		{name: "only digits", pwd: "12345678", wantErr: "at least 2"},
		{name: "only lowercase", pwd: "abcdefgh", wantErr: "at least 2"},
		{name: "only uppercase", pwd: "ABCDEFGH", wantErr: "at least 2"},
		{name: "only specials", pwd: "!@#$%^&*", wantErr: "at least 2"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePassword(tc.pwd)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("validatePassword(%q) unexpected error: %v", tc.pwd, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validatePassword(%q) returned nil, want error containing %q", tc.pwd, tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("validatePassword(%q) = %v, want substring %q", tc.pwd, err, tc.wantErr)
			}
		})
	}
}

func TestCreateRejectsBadPasswordBeforeRequest(t *testing.T) {
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
			ctx, form, cleanup := setupCreateMock(t)
			defer cleanup()

			cmd := newCreate(ctx)
			args := requiredCreateArgs()
			for i := range args {
				if args[i] == "--password" {
					args[i+1] = tc.pwd
					break
				}
			}
			cmd.SetArgs(args)
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
			if len(*form) != 0 {
				t.Fatal("invalid password must not reach the API")
			}
		})
	}
}
