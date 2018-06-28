package circuitbreaker

import (
	"context"
	"testing"

	"github.com/go-kit/kit/endpoint"
	"github.com/sony/gobreaker"
)

type retryableError struct{}

func (_ *retryableError) Error() string   { return "error" }
func (_ *retryableError) Retryable() bool { return true }

type nonRetryableError struct{}

func (_ *nonRetryableError) Error() string   { return "error" }
func (_ *nonRetryableError) Retryable() bool { return false }

func TestCircuitBreaker(t *testing.T) {
	{
		cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{ReadyToTrip: func(_ gobreaker.Counts) bool { return true }})
		called := 0
		e := NewCircuitBreaker(cb)(mkFailingEndpoint(&called, &retryableError{}))
		_, err := e(context.TODO(), nil)
		if err == nil {
			t.Error("the circuit breaker should return an error")
		}
		_, err = e(context.TODO(), nil)
		if err == nil {
			t.Error("the circuit breaker should return an error")
		}
		if called > 1 {
			t.Error("retryable errors should trigger the circuit breaker")
		}
	}
	{
		cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{ReadyToTrip: func(_ gobreaker.Counts) bool { return true }})
		called := 0
		e := NewCircuitBreaker(cb)(mkFailingEndpoint(&called, &nonRetryableError{}))
		_, err := e(context.TODO(), nil)
		if err == nil {
			t.Error("the circuit breaker should return an error")
		}
		_, err = e(context.TODO(), nil)
		if err == nil {
			t.Error("the circuit breaker should return an error")
		}
		if called <= 1 {
			t.Error("non retryable errors should not trigger the circuit breaker")
		}
	}
}

func mkFailingEndpoint(count *int, err error) endpoint.Endpoint {
	return func(_ context.Context, _ interface{}) (interface{}, error) {
		*count++
		return nil, err
	}
}
