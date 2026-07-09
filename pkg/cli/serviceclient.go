package cli

import (
	ucloud "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

// NewServiceClient returns an authed SDK service client for the active profile.
// The host injects the credential + handler path, so oauth and AK/SK profiles
// still share one code path (§9: no auth regression). ctor is e.g. udb.NewClient.
//
// Go methods cannot have type parameters, so this is a package-level generic
// function rather than a *Context method.
func NewServiceClient[T ucloud.ServiceClient](ctx *Context, ctor func(*ucloud.Config, *auth.Credential) T) T {
	if ctx == nil || ctx.clientConfig == nil || ctx.buildCredential == nil || ctx.attachHandlers == nil {
		panic("cli.NewServiceClient called without service-client dependencies")
	}
	c := ctor(ctx.clientConfig(), ctx.buildCredential())
	ctx.attachHandlers(c)
	return c
}
