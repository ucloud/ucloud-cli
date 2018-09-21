package trace

// TraceInfo is the detail information of sdk invoking.
type TraceInfo interface {
	GetSDKVersion() string

	GetSDKRequest() interface{}
	SetSDKRequest(data interface{}) error

	GetSDKResponse() interface{}
	SetSDKResponse(data interface{}) error

	IsError() bool
	SetError(error) error

	GetTraceback() []StacktraceFrame

	SetExtraData(key string, val interface{}) error
	GetExtra() map[string]interface{}
}

// Tracer is used to send trace information
type Tracer interface {
	Send(TraceInfo, map[string]string) error
}

// type AggTracer interface {
// 	TraceInfo
// 	Send() error
// }
