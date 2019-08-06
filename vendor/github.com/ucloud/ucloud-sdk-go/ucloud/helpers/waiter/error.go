package waiter

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	errTimeoutConf = errors.New("timeout cannot be set zero")
)

// TimeoutError is returned when WaitForState times out
type TimeoutError struct {
	LastError      error
	LastState      string
	Timeout        time.Duration
	ExpectedStates []string
}

func (e *TimeoutError) Error() string {
	errors := []string{"cannot waiting for resource is completed"}

	if e.Timeout > 0 {
		errors = append(errors, fmt.Sprintf("timeout: %s", e.Timeout))
	}

	if e.LastState != "" {
		errors = append(errors, fmt.Sprintf("last state: %q", e.LastState))
	}

	if len(e.ExpectedStates) > 0 {
		errors = append(errors, fmt.Sprintf("want: %q", strings.Join(e.ExpectedStates, ",")))
	}

	if e.LastError != nil {
		errors = append(errors, fmt.Sprintf("err: %s", e.LastError))
	}

	return strings.Join(errors, ", ")
}
