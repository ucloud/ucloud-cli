package pathx

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type PathXClient struct {
	client *ucloud.Client
}

func NewClient(config *ucloud.Config, credential *auth.Credential) *PathXClient {
	client := ucloud.NewClient(config, credential)
	return &PathXClient{
		client: client,
	}
}
