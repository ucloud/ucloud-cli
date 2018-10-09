package request

import "time"

// Common is the common request
type Common interface {
	GetAction() string
	SetAction(string) error

	GetRegion() string
	SetRegion(string) error

	GetProjectId() string
	SetProjectId(string) error

	SetRetryCount(int)
	GetRetryCount() int

	WithRetry(int)
	GetMaxretries() int

	WithTimeout(time.Duration)
	GetTimeout() time.Duration

	SetRequestTime(time.Time)
	GetRequestTime() time.Time
}

// CommonBase is the base struct of common request
type CommonBase struct {
	Action    *string
	Region    *string
	ProjectId *string

	maxRetries  int
	retryCount  int
	timeout     time.Duration
	requestTIme time.Time
}

// SetRetryCount will set retry count of request
func (c *CommonBase) SetRetryCount(retryCount int) {
	c.retryCount = retryCount
}

// GetRetryCount will return retry count of request
func (c *CommonBase) GetRetryCount() int {
	return c.retryCount
}

// WithRetry will set max retry count of request
func (c *CommonBase) WithRetry(maxRetries int) {
	c.maxRetries = maxRetries
}

// GetMaxretries will return max retry count of request
func (c *CommonBase) GetMaxretries() int {
	return c.maxRetries
}

// WithTimeout will set timeout of request
func (c *CommonBase) WithTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// GetTimeout will get timeout of request
func (c *CommonBase) GetTimeout() time.Duration {
	return c.timeout
}

// SetRequestTime will set timeout of request
func (c *CommonBase) SetRequestTime(requestTIme time.Time) {
	c.requestTIme = requestTIme
}

// GetRequestTime will get timeout of request
func (c *CommonBase) GetRequestTime() time.Time {
	return c.requestTIme
}

// GetAction will return action of request
func (c *CommonBase) GetAction() string {
	return *c.Action
}

// SetAction will set action of request
func (c *CommonBase) SetAction(val string) error {
	c.Action = &val
	return nil
}

// GetRegion will return region of request
func (c *CommonBase) GetRegion() string {
	if c.Region == nil {
		return ""
	}
	return *c.Region
}

// SetRegion will set region of request
func (c *CommonBase) SetRegion(val string) error {
	c.Region = &val
	return nil
}

// GetProjectId will get project id of request
func (c *CommonBase) GetProjectId() string {
	if c.ProjectId == nil {
		return ""
	}
	return *c.ProjectId
}

// SetProjectId will set project id of request
func (c *CommonBase) SetProjectId(val string) error {
	c.ProjectId = &val
	return nil
}
