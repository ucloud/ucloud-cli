package trace

import (
	"github.com/ucloud/ucloud-sdk-go/sdk/version"
)

// DasTraceInfo is an implementation of DasTraceInfo
type DasTraceInfo struct {
	SDKVersion string
	Channel    string

	request  interface{}
	response interface{}

	err       error
	traceback []*StacktraceFrame
	extra     map[string]interface{}
}

// NewDasTraceInfo will create a new trace info struct with default information
func NewDasTraceInfo() DasTraceInfo {
	return DasTraceInfo{
		Channel:    "ucloud",
		SDKVersion: version.Version,
		extra:      make(map[string]interface{}),
	}
}

// GetSDKVersion will return version of sdk
func (d *DasTraceInfo) GetSDKVersion() string {
	return d.SDKVersion
}

// GetSDKRequest will return sdk request data
func (d *DasTraceInfo) GetSDKRequest() interface{} {
	return d.request
}

// SetSDKRequest will set sdk request data
func (d *DasTraceInfo) SetSDKRequest(data interface{}) error {
	d.request = data
	return nil
}

// GetSDKResponse will return sdk request data
func (d *DasTraceInfo) GetSDKResponse() interface{} {
	return d.response
}

// SetSDKResponse will set sdk request data
func (d *DasTraceInfo) SetSDKResponse(data interface{}) error {
	d.response = data
	return nil
}

// IsError will return if this trace record is error
func (d *DasTraceInfo) IsError() bool {
	return d.err != nil
}

// SetError wil set record with error
// if the error has traceback, it will save traceback into this trace infomation record.
func (d *DasTraceInfo) SetError(err error) error {
	// TODO: capture error with stacktrace
	d.err = err
	return nil
}

// GetTraceback will return all frames of stacktrace
// See also "github.com/pkg/errors"
func (d *DasTraceInfo) GetTraceback() []StacktraceFrame {
	return []StacktraceFrame{}
}

// SetExtraData will set some extra data will be sent to remote server
func (d *DasTraceInfo) SetExtraData(key string, val interface{}) error {
	if d.extra == nil {
		d.extra = make(map[string]interface{})
	}

	d.extra[key] = val
	return nil
}

// GetExtra will get the key-value map of extra data
func (d *DasTraceInfo) GetExtra() map[string]interface{} {
	return d.extra
}
