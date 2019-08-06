package pathx

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// PathXClient is the client of PathX
type PathXClient struct {
	*ucloud.Client
}

// NewClient will return a instance of PathXClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *PathXClient {
	client := ucloud.NewClient(config, credential)
	return &PathXClient{
		client,
	}
}
