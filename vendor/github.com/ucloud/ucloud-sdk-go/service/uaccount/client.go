package uaccount

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/auth"
)

type UAccountClient struct {
	client *sdk.Client
}

func NewClient(config *sdk.ClientConfig, credential *auth.Credential) *UAccountClient {
	client := sdk.NewClient(config, credential)
	return &UAccountClient{
		client: client,
	}
}
