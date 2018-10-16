package vpc

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type VPCClient struct {
	client *ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *VPCClient {
	client := ucloud.NewClient(config, credential)
	return &VPCClient{
		client: client,
	}
}
