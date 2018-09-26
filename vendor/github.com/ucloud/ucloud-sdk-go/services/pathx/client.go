package pathx

import (
	"github.com/ucloud/ucloud-sdk-go/sdk"
	"github.com/ucloud/ucloud-sdk-go/sdk/auth"
)

type PathXClient struct {
	client *sdk.Client
}

func NewClient(config *sdk.Config, credential *auth.Credential) *PathXClient {
	client := sdk.NewClient(config, credential)
	return &PathXClient{
		client: client,
	}
}
