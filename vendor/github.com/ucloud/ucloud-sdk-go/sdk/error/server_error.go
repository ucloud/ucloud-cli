package error

import (
	"errors"
)

var (
	InternalError = errors.New("server.InternalError")
)
