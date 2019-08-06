package udb

import (
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// UDBClient is the client of UDB
type UDBClient struct {
	*ucloud.Client
}

// NewClient will return a instance of UDBClient
func NewClient(config *ucloud.Config, credential *auth.Credential) *UDBClient {
	client := ucloud.NewClient(config, credential)
	return &UDBClient{
		client,
	}
}
