package kitty

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-kit/kit/endpoint"
)

func TestEndpointResponseEncode(t *testing.T) {
	cfg := Config{}
	HTTPTransport := NewHTTPTransport(cfg)
	HTTPTransport.Endpoint("GET", "/test/default", func(ctx context.Context, r interface{}) (interface{}, error) {
		return "default response", nil
	}).Endpoint("GET", "/test/override", func(ctx context.Context, r interface{}) (interface{}, error) {
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
		})
		if rec.Code != 200 {
			t.Errorf("default HTTP response status expected: %d", rec.Code)
		}
		body := rec.Body.String()
		if strings.TrimSpace(body) != `"default response"` {
			t.Errorf("different body expected: %s", body)
		}
	}
	{
		rec := httptest.NewRecorder()
		HTTPTransport.ServeHTTP(rec, &http.Request{
			Method:     "GET",
			RequestURI: "/test/override",
			URL: &url.URL{
				Path: "/test/override",
			},
		})
		if rec.Code != 501 {
			t.Errorf("override HTTP response status expected: %d", rec.Code)
		}
		body := rec.Body.String()
		if body != "" {
			t.Errorf("different body expected: %s", body)
		}
	}
}

func TestDefaultResponseEncode(t *testing.T) {
	cfg := Config{
		EncodeResponse: func(ctx context.Context, w http.ResponseWriter, r interface{}) error {
			w.WriteHeader(501)
			w.Write([]byte("response:"))
			return json.NewEncoder(w).Encode(r)
		},
	}
	HTTPTransport := NewHTTPTransport(cfg).
		Endpoint("GET", "/test", func(ctx context.Context, r interface{}) (interface{}, error) {
			return "default response", nil
		})
	err := HTTPTransport.RegisterEndpoints(func(e endpoint.Endpoint) endpoint.Endpoint {
		return e
	})
	if err != nil {
		t.Errorf("error occurred: %+v", err)
	}
	rec := httptest.NewRecorder()
	HTTPTransport.ServeHTTP(rec, &http.Request{
		Method:     "GET",
		RequestURI: "/test",
		URL: &url.URL{
			Path: "/test",
		},
	})
	if rec.Code != 501 {
		t.Errorf("default HTTP response status expected: %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.TrimSpace(body) != `response:"default response"` {
		t.Errorf("different body expected: %s", body)
	}
}
