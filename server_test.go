package kitty

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/kit/endpoint"
)

func TestServer(t *testing.T) {
	shutdownCalled := false
	ctx, cancel := context.WithCancel(context.TODO())
	exitError := make(chan error)
	tr := NewHTTPTransport(DefaultConfig).
		Endpoint("POST", "/foo", testEP, Decoder(goodDecoder)).
		Endpoint("GET", "/decoding_error", testEP, Decoder(badDecoder))
	srv := NewServer(tr).Shutdown(func() {
		shutdownCalled = true
	})
	go func() {
		exitError <- srv.Run(ctx)
	}()

	start := time.Now()
	for {
		resp, err := http.Get("http://localhost:8080/alivez")
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if time.Since(start) > 500*time.Millisecond {
			t.Fatal("server did not start within 500msec or liveness returned an error")
		}
		time.Sleep(50 * time.Millisecond)
	}

	{
		resp, err := http.Get("http://localhost:8080/readyz")
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Error("readyness returned an error")
		}
	}

	{
		resp, err := http.Post("http://localhost:8080/foo", "application/json", bytes.NewBufferString(`{"foo":"bar"}`))
		if err != nil {
			t.Errorf("http.Get returned an error : %s", err)
		} else {
			if resp.StatusCode != 200 {
				t.Errorf("receive a %d status instead of 200", resp.StatusCode)
			}
			resData := testStruct{}
			err := json.NewDecoder(resp.Body).Decode(&resData)
			resp.Body.Close()
			if err != nil {
				t.Errorf("json.Decode returned an error : %s", err)
			} else if !reflect.DeepEqual(resData, testStruct{Foo: "bar"}) {
				t.Errorf("http.Get returned invalid data : %+v", resData)
			}
		}
	}

	{
		resp, err := http.Post("http://localhost:8080/foo", "application/json", bytes.NewBufferString(`{"status":404}`))
		if err != nil {
			t.Errorf("http.Get returned an error : %s", err)
		} else {
			if resp.StatusCode != 404 {
				t.Errorf("receive a %d status instead of 404", resp.StatusCode)
			}
		}
	}

	{
		resp, err := http.Get("http://localhost:8080/decoding_error")
		if err != nil {
			t.Errorf("http.Get returned an error : %s", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("A decoding error should return a BadRequest status, not %d", resp.StatusCode)
			}
		}
	}

	cancel()
	select {
	case <-time.After(time.Second):
		t.Error("Server.Run has not stopped after 1sec")
	case err := <-exitError:
		if err != nil && err != context.Canceled {
			t.Errorf("Server.Run returned an error : %s", err)
		}
	}
	if !shutdownCalled {
		t.Error("Shutdown functions are not called")
	}
}

type testStruct struct {
	Foo    string `json:"foo"`
	Status int    `json:"status"`
}

func testEP(_ context.Context, req interface{}) (interface{}, error) {
	if r, ok := req.(*testStruct); ok && r.Status != 0 {
		return nil, httpError(r.Status)
	}
	return req, nil
}

func goodDecoder(_ context.Context, r *http.Request) (interface{}, error) {
	request := &testStruct{}
	err := json.NewDecoder(r.Body).Decode(request)
	return request, err
}

func badDecoder(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, errors.New("decoding error")
}

type failingTransport struct {
	Transport
}

func (*failingTransport) RegisterEndpoints(m endpoint.Middleware) error { return nil }
func (*failingTransport) Start(ctx context.Context) error               { return errors.New("unable to start") }
func (*failingTransport) Shutdown(ctx context.Context) error            { return nil }

type workingTransport struct {
	running chan struct{}
	Transport
}

func (*workingTransport) RegisterEndpoints(m endpoint.Middleware) error { return nil }
func (t *workingTransport) Start(ctx context.Context) error {
	t.running = make(chan struct{})
	defer func() {
		close(t.running)
	}()
	<-ctx.Done()
	return nil
}
func (*workingTransport) Shutdown(ctx context.Context) error { return nil }

func TestStartError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	exitError := make(chan error)
	wt := &workingTransport{}
	srv := NewServer(&failingTransport{}, wt)
	go func() {
		exitError <- srv.Run(ctx)
	}()

	select {
	case <-ctx.Done():
		t.Error("Server.Run has not stopped after 1sec")
	case err := <-exitError:
		if err == nil || err.Error() != "unable to start" {
			t.Errorf("Server.Run returned an invalid error : %v", err)
		}
	}
	select {
	case <-ctx.Done():
		t.Error("Alternate transport has not stopped after 1sec")
	case <-wt.running:
	}
}
