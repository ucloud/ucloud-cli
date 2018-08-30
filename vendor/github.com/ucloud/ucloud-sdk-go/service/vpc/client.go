package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/auth"
)

type VPCClient struct {
	client *sdk.Client
}

func NewClient(config *sdk.ClientConfig, credential *auth.Credential) *VPCClient {
	client := sdk.NewClient(config, credential)
	return &VPCClient{
		client: client,
	}
}
