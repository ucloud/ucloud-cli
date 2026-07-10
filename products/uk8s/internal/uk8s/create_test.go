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

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/pkg/cli"
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

	oldClientConfig, oldCredential, oldConfig := base.ClientConfig, base.AuthCredential, base.ConfigIns
	base.ClientConfig = &sdk.Config{
		Region:    "cn-sh2",
		ProjectId: "org-test",
		BaseUrl:   server.URL,
	}
	base.AuthCredential = &base.CredentialConfig{PublicKey: "public", PrivateKey: "private"}
	base.ConfigIns = &base.AggConfig{Region: "cn-sh2", ProjectID: "org-test", BaseURL: server.URL}

	var out, errOut bytes.Buffer
	ctx := cli.NewContext(cli.Deps{
		In:     strings.NewReader(""),
		Out:    &out,
		Err:    &errOut,
		Format: cli.OutputTable,
		Config: base.ConfigIns,
	})
	cleanup := func() {
		server.Close()
		base.ClientConfig, base.AuthCredential, base.ConfigIns = oldClientConfig, oldCredential, oldConfig
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
		"--async",
	}
}

func TestCreateRequestMatchesDocument(t *testing.T) {
	ctx, form, cleanup := setupCreateMock(t)
	defer cleanup()

	cmd := newCreate(ctx)
	args := append(requiredCreateArgs(),
		"--k8s-version", "1.34.5",
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
		"K8sVersion", "ChargeType", "Quantity", "UserData", "InitScript",
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
