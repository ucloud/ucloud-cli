package cmd

import (
	"fmt"

	"github.com/ucloud/ucloud-cli/base"
	"github.com/ucloud/ucloud-cli/pkg/command"
	sdk "github.com/ucloud/ucloud-sdk-go/ucloud"
	"github.com/ucloud/ucloud-sdk-go/ucloud/auth"
)

type runtimeState struct {
	Configs    *base.AggConfigManager
	Config     *base.AggConfig
	SDKConfig  *sdk.Config
	Credential *base.CredentialConfig
}

var activeRuntime *runtimeState
var runtimeAutoStub = true

func buildRuntimeFromBaseGlobals() *runtimeState {
	return &runtimeState{
		Configs:    base.AggConfigListIns,
		Config:     base.ConfigIns,
		SDKConfig:  base.ClientConfig,
		Credential: base.AuthCredential,
	}
}

func ensureRuntime() *runtimeState {
	if activeRuntime == nil {
		activeRuntime = buildRuntimeFromBaseGlobals()
	}
	if runtimeAutoStub && activeRuntime.SDKConfig == nil {
		activeRuntime.SDKConfig = &sdk.Config{BaseUrl: base.DefaultBaseURL}
		base.ClientConfig = activeRuntime.SDKConfig
	}
	if runtimeAutoStub && activeRuntime.Credential == nil {
		activeRuntime.Credential = &base.CredentialConfig{}
		base.AuthCredential = activeRuntime.Credential
	}
	return activeRuntime
}

func setActiveRuntimeFromBaseGlobals() {
	runtimeAutoStub = true
	activeRuntime = buildRuntimeFromBaseGlobals()
}

func runtimeDefaults() command.Defaults {
	rt := ensureRuntime()
	if rt == nil || rt.Config == nil {
		return command.Defaults{}
	}
	return command.Defaults{Region: rt.Config.Region, Zone: rt.Config.Zone, ProjectID: rt.Config.ProjectID}
}

func runtimeClientConfig() *sdk.Config {
	rt := ensureRuntime()
	if rt == nil {
		return nil
	}
	if !runtimeAutoStub && rt.SDKConfig == nil {
		panic("cmd runtime disabled for snapshot completion")
	}
	return rt.SDKConfig
}

func runtimeCredential() *auth.Credential {
	rt := ensureRuntime()
	if rt == nil {
		return base.BuildCredentialFrom(nil)
	}
	return base.BuildCredentialFrom(rt.Credential)
}

func attachRuntimeHandlers(sc sdk.ServiceClient) {
	rt := ensureRuntime()
	if rt == nil {
		base.AttachHandlersWith(sc, nil, nil, nil)
		return
	}
	base.AttachHandlersWith(sc, rt.Credential, rt.Config, rt.Configs)
}

func newServiceClient[T sdk.ServiceClient](ctor func(*sdk.Config, *auth.Credential) T) T {
	rt := ensureRuntime()
	if rt == nil || rt.SDKConfig == nil {
		panic("cmd runtime is not initialized")
	}
	client := ctor(rt.SDKConfig, base.BuildCredentialFrom(rt.Credential))
	base.AttachHandlersWith(client, rt.Credential, rt.Config, rt.Configs)
	return client
}

func newServiceClientForConfig[T sdk.ServiceClient](cfg *base.AggConfig, ctor func(*sdk.Config, *auth.Credential) T) (T, error) {
	var zero T
	sdkConfig, credConfig, err := base.BuildClientRuntime(cfg)
	if sdkConfig == nil {
		return zero, fmt.Errorf("build sdk config failed")
	}
	client := ctor(sdkConfig, base.BuildCredentialFrom(credConfig))
	rt := ensureRuntime()
	var manager *base.AggConfigManager
	if rt != nil {
		manager = rt.Configs
	}
	base.AttachHandlersWith(client, credConfig, cfg, manager)
	return client, err
}
