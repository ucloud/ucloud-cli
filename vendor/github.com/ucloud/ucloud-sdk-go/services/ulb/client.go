package ulb

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/auth"
)

type ULBClient struct {
	client *sdk.Client
}

func NewClient(config *sdk.Config, credential *auth.Credential) *ULBClient {
	client := sdk.NewClient(config, credential)
	return &ULBClient{
		client: client,
	}
}
