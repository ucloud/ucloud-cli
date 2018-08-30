package unet

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/auth"
)

type UNetClient struct {
	client *sdk.Client
}

func NewClient(config *sdk.ClientConfig, credential *auth.Credential) *UNetClient {
	client := sdk.NewClient(config, credential)
	return &UNetClient{
		client: client,
	}
}
