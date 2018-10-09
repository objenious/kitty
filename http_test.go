package kitty

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-kit/kit/endpoint"

	kithttp "github.com/go-kit/kit/transport/http"
)

type Response struct {
	status string
}

func TestEndpointResponseEncode(t *testing.T) {
	cfg := Config{}
	HTTPTransport := NewHTTPTransport(cfg)
	defaultCalled := false
	overrideCalled := false
	HTTPTransport.Endpoint("GET", "/test/default", func(ctx context.Context, r interface{}) (interface{}, error) {
		defaultCalled = true
		return "default response", nil
	}).Endpoint("GET", "/test/override", func(ctx context.Context, r interface{}) (interface{}, error) {
		overrideCalled = true
		return "override response", nil
	}, Encoder(func(ctx context.Context, w http.ResponseWriter, r interface{}) error {
		w.WriteHeader(501)
		return nil
	}))
	HTTPTransport.RegisterEndpoints(func(e endpoint.Endpoint) endpoint.Endpoint {
		return e
	})
	{
		rec := httptest.NewRecorder()
		HTTPTransport.ServeHTTP(rec, &http.Request{
			Method:     "GET",
			RequestURI: "/test/default",
			URL: &url.URL{
				Path: "/test/default",
			},
			Body: ioutil.NopCloser(bytes.NewBuffer([]byte(`{}`))),
		})
		if !defaultCalled {
			t.Error("default endpoint not called")
		}
		if rec.Code != 200 {
			t.Errorf("default HTTP response status expected: %d", rec.Code)
		}
		body := string(rec.Body.Bytes())
		if strings.TrimSpace(body) != `"default response"` {
			t.Errorf("different body expected: %s", body)
		}
	}
	overrideCalled = false
	defaultCalled = false
	{
		rec := httptest.NewRecorder()
		HTTPTransport.ServeHTTP(rec, &http.Request{
			Method:     "GET",
			RequestURI: "/test/override",
			URL: &url.URL{
				Path: "/test/override",
			},
			Body: ioutil.NopCloser(bytes.NewBuffer([]byte("OK:override"))),
		})
		if !overrideCalled {
			t.Error("override endpoint not called")
		}
		if rec.Code != 501 {
			t.Errorf("override HTTP response status expected: %d", rec.Code)
		}
		body := string(rec.Body.Bytes())
		if body != "" {
			t.Errorf("different body expected: %s", body)
		}
	}
}

func TestDefaultResponseEncode(t *testing.T) {
	cfg := Config{
		EncodeResponse: func(ctx context.Context, w http.ResponseWriter, r interface{}) error {
			w.WriteHeader(501)
			return nil
		},
	}

	defaultCalled := false
	overrideCalled := false
	HTTPTransport := NewHTTPTransport(cfg).
		Endpoint("GET", "/test/override", func(ctx context.Context, r interface{}) (interface{}, error) {
			overrideCalled = true
			return "override response", nil
		}, Encoder(kithttp.EncodeJSONResponse)).
		Endpoint("GET", "/test/default", func(ctx context.Context, r interface{}) (interface{}, error) {
			defaultCalled = true
			return "default response", nil
		})
	err := HTTPTransport.RegisterEndpoints(func(e endpoint.Endpoint) endpoint.Endpoint {
		return e
	})
	if err != nil {
		t.Errorf("error occurred: %+v", err)
	}
	{
		rec := httptest.NewRecorder()
		HTTPTransport.ServeHTTP(rec, &http.Request{
			Method:     "GET",
			RequestURI: "/test/default",
			URL: &url.URL{
				Path: "/test/default",
			},
			Body: ioutil.NopCloser(bytes.NewBuffer([]byte(`{}`))),
		})
		if !defaultCalled {
			t.Error("default endpoint not called")
		}
		if rec.Code != 501 {
			t.Errorf("default HTTP response status expected: %d", rec.Code)
		}
		body := string(rec.Body.Bytes())
		if body != "" {
			t.Errorf("different body expected: %s", body)
		}
	}
	overrideCalled = false
	defaultCalled = false
	{
		rec := httptest.NewRecorder()
		HTTPTransport.ServeHTTP(rec, &http.Request{
			Method:     "GET",
			RequestURI: "/test/override",
			URL: &url.URL{
				Path: "/test/override",
			},
			Body: ioutil.NopCloser(bytes.NewBuffer([]byte("OK:override"))),
		})
		if !overrideCalled {
			t.Error("override endpoint not called")
		}
		if rec.Code != 200 {
			t.Errorf("override HTTP response status expected: %d", rec.Code)
		}
		body := string(rec.Body.Bytes())
		if strings.TrimSpace(body) != `"override response"` {
			t.Errorf("different body expected: %s", body)
		}
	}
}
