package backoff

import (
	"context"
	"testing"

	"github.com/cenkalti/backoff"
	"github.com/go-kit/kit/endpoint"
)

type retryableError struct{}

func (*retryableError) Error() string   { return "error" }
func (*retryableError) Retryable() bool { return true }

type nonRetryableError struct{}

func (*nonRetryableError) Error() string   { return "error" }
func (*nonRetryableError) Retryable() bool { return false }

func TestBackoff(t *testing.T) {
	bo := backoff.NewExponentialBackOff()
	{
		e := NewBackoff(bo)(mkFailingEndpoint(&retryableError{}, &retryableError{}))
		res, err := e(context.TODO(), nil)
		if err != nil {
			t.Error("With a retryable error, backoff should not return an error")
		}
		if res != "OK" {
			t.Error("With a retryable error, backoff should have returned the right result")
		}
	}
	{
		e := NewBackoff(bo)(mkFailingEndpoint(&nonRetryableError{}))
		res, err := e(context.TODO(), nil)
		if err == nil {
			t.Error("With a non retryable error, backoff should return an error")
		}
		if res != nil {
			t.Error("With a non retryable error, backoff should have returned no result")
		}
	}

}

func mkFailingEndpoint(errors ...error) endpoint.Endpoint {
	return func(_ context.Context, _ interface{}) (interface{}, error) {
		if len(errors) == 0 {
			return "OK", nil
		}
		err := errors[0]
		errors = errors[1:]
		return nil, err
	}
}
