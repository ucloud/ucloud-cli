package response

// Common describe a response of action,
// it is only used for ucloud open api v1 via HTTP GET and action parameters.
type Common interface {
	GetRetCode() int
	GetMessage() string
	GetAction() string
}

// CommonBase has common attribute and method,
// it also implement ActionResponse interface.
type CommonBase struct {
	Action  string
	RetCode int
	Message string
}

// GetRetCode will return the error code of ucloud api
// Error is non-zero and succuess is zero
func (c *CommonBase) GetRetCode() int {
	return c.RetCode
}

// GetMessage will return the error message of ucloud api
func (c *CommonBase) GetMessage() string {
	return c.Message
}

// GetAction will return the request action of ucloud api
func (c *CommonBase) GetAction() string {
	return c.Action
}
