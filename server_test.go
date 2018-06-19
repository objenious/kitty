package kitty

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.TODO())
	exitError := make(chan error)
	srv := NewServer().HTTPEndpoint(testEP, Path("/foo"), Method("GET"))
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
		resp, err := http.Get("http://localhost:8080/foo")
		if err != nil {
			t.Errorf("http.Get returned an error : %s", err)
		} else {
			defer resp.Body.Close()
			resData := testStruct{}
			err := json.NewDecoder(resp.Body).Decode(&resData)
			if err != nil {
				t.Errorf("json.Decode returned an error : %s", err)
			} else if !reflect.DeepEqual(resData, testData) {
				t.Errorf("http.Get returned invalid data : %+v", resData)
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
}

type testStruct struct {
	Foo string
}

var testData = testStruct{Foo: "bar"}

func testEP(_ context.Context, _ interface{}) (interface{}, error) {
	return testData, nil
}
