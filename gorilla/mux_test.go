package gorilla

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/objenious/kitty"
)

func TestRouter(t *testing.T) {
	notfoundcalled := false
	ctx, cancel := context.WithCancel(context.TODO())
	tr := kitty.NewHTTPTransport(kitty.DefaultConfig).
		Endpoint("POST", "/foo", testEP, kitty.Decoder(goodDecoder)).
		Router(Router(), kitty.NotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			notfoundcalled = true
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, "OK")
		})))
	srv := kitty.NewServer(tr)
	go func() {
		srv.Run(ctx)
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
		resp, err := http.Get("http://localhost:8080/does_not_exist")
		if err != nil {
			t.Errorf("http.Get returned an error : %s", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("A call to an unknown url should return a StatusNotFound status, not %d", resp.StatusCode)
			}
			if !notfoundcalled {
				t.Error("the not found handler was not called")
			}
		}
	}

	cancel()
}

type testStruct struct {
	Foo string `json:"foo"`
}

func testEP(_ context.Context, req interface{}) (interface{}, error) {
	return req, nil
}

func goodDecoder(_ context.Context, r *http.Request) (interface{}, error) {
	request := &testStruct{}
	err := json.NewDecoder(r.Body).Decode(request)
	return request, err
}
