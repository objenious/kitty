package kitty

import (
	"errors"
	"net/http"
	"testing"
)

func TestRetryable(t *testing.T) {
	{
		err := errors.New("foo")
		err2 := Retryable(err)
		if err.Error() != err2.Error() {
			t.Error("Retryable should not change the error text")
		}
		if !IsRetryable(err2) {
			t.Error("Retryable should generate a retryable error")
		}
	}
	{
		err := httpError(http.StatusBadRequest)
		err2 := Retryable(err)
		if !IsRetryable(err2) {
			t.Error("Retryable should transform a non retryable error into a retryable error")
		}
	}
}
