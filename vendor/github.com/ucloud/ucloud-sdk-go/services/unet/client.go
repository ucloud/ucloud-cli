package unet

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type UNetClient struct {
	client *ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *UNetClient {
	client := ucloud.NewClient(config, credential)
	return &UNetClient{
		client: client,
	}
}
