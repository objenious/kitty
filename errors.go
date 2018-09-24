package kitty

import "net/http"

// Retryabler defines an error that may be temporary. A function returning a retryable error may be executed again.
type Retryabler interface {
	Retryable() bool
}

// IsRetryable checks if an error is retryable (i.e. implements Retryabler and Retryable returns true).
// Retryable errors may be wrapped using github.com/pkg/errors.
// If the error is nil or does not implement Retryabler, false is returned.
func IsRetryable(err error) bool {
	type causer interface {
		Cause() error
	}

	for err != nil {
		if retry, ok := err.(Retryabler); ok {
			return retry.Retryable()
		}
		cause, ok := err.(causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return false
}

type retryableError struct {
	error
}

func (retryableError) Retryable() bool {
	return true
}

var _ error = retryableError{}
var _ Retryabler = retryableError{}

// Retryable defines an error as retryable.
func Retryable(err error) error {
	return retryableError{error: err}
}

// HTTPError builds an error based on a http.Response. If status code is < 300 or 304, nil is returned.
// 429, 5XX errors are Retryable.
func HTTPError(resp *http.Response) error {
	if resp.StatusCode < 300 || resp.StatusCode == http.StatusNotModified {
		return nil
	}
	return httpError(resp.StatusCode)
}

type httpError int

func (err httpError) Error() string {
	switch err {
	case 429:
		return "Too Many Requests"
	default:
		return http.StatusText(int(err))
	}
}

func (err httpError) StatusCode() int {
	return int(err)
}

func (err httpError) Retryable() bool {
	switch int(err) {
	case http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusServiceUnavailable, http.StatusInternalServerError:
		return true
	case 429:
		return true
	default:
		return false
	}
}
