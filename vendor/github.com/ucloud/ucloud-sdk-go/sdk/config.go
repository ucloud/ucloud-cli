package sdk

import (
	"time"

	"github.com/Sirupsen/logrus"
)

type ClientConfig struct {
	// Region is the region of backend service
	// See also <doc link> ...
	Region string

	// ProjectId is the unique identify of project, used for organize resources,
	// Most of resources should belong to a project.
	// Sub-Account must have an project id.
	// See also <doc link>
	ProjectId string

	// BaseUrl is the url of backend api
	// See also <doc link> ...
	BaseUrl string

	// Logger and LogLevel is the configuration of logrus,
	// if logger not be set, use standard output with json formatter as default,
	// if logLevel not be set, use INFO level as default.
	Logger   *logrus.Logger
	LogLevel logrus.Level

	// Timeout is timeout for every request.
	Timeout time.Duration

	// Trace will invoke when any request is completed.
	// Tracer     utrace.Tracer
	// TracerData map[string]interface{}

	// UserAgent is an attribute for sdk client, used for distinguish who using sdk.
	// See also https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent
	// It will append to the end of sdk user-agent.
	// eg. "Terraform/0.10.1" -> "GO/1.9.1 GO-SDK/0.1.0 Terraform/0.10.1"
	// warn. it will conflict with the User-Agent of HTTPHeaders
	UserAgent string

	// HTTPHeaders is the specific headers sent to remote server via http protocal
	// It is avaliabled when http protocal is enabled.
	// HTTPHeaders map[string]string

	// AutoRetry is a switch to enable retry policy for timeout/connect failing
	// if AutoRetry is enabled, it will enable default retry policy using exponential backoff.
	AutoRetry bool

	// MaxRetries is the number of max retry times.
	MaxRetries int

	// AllowTrace 是否允许上报数据
	AllowTrace bool
}
