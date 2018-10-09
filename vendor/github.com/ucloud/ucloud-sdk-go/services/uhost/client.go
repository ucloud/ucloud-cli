package uhost

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/auth"
)

type UHostClient struct {
	client *sdk.Client
}

func NewClient(config *sdk.Config, credential *auth.Credential) *UHostClient {
	client := sdk.NewClient(config, credential)
	return &UHostClient{
		client: client,
	}
}
