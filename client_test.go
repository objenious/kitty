package kitty

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	kithttp "github.com/go-kit/kit/transport/http"
)

func TestClient(t *testing.T) {
	h := testHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()

	client := ts.Client()
	u, _ := url.Parse(ts.URL)
	e := NewClient("GET", u, kithttp.EncodeJSONRequest, decodeTestResponse, kithttp.SetClient(client)).Endpoint()
	h.statuses = []int{http.StatusServiceUnavailable}
	_, err := e(context.TODO(), nil)
	if err == nil {
		t.Error("When calling a failed server, the client should return an error")
	}
	if !IsRetryable(err) {
		t.Error("The returned error for http.StatusServiceUnavailable should be retryable")
	}
	h.statuses = []int{http.StatusBadRequest}
	_, err = e(context.TODO(), nil)
	if err == nil {
		t.Error("When calling a failed server, the client should return an error")
	}
	if IsRetryable(err) {
		t.Error("The returned error for http.StatusBadRequest should not be retryable")
	}
	h.statuses = []int{}
	res, err := e(context.TODO(), nil)
	if err != nil {
		t.Errorf("When calling a working server, the client should not return an error, got %s", err)
	} else if !reflect.DeepEqual(res, testData) {
		t.Errorf("The endpoint returned invalid data : %+v", res)
	}
}

func TestClientWithError(t *testing.T) {
	h := testHandler{}
	ts := httptest.NewServer(&h)
	defer ts.Close()

	client := ts.Client()
	u, _ := url.Parse(ts.URL)
	e := NewClientWithError("GET", u, kithttp.EncodeJSONRequest, decodeTestResponse, kithttp.SetClient(client)).Endpoint()

	h.statuses = []int{http.StatusServiceUnavailable}
	resp, err := e(context.TODO(), nil)
	if err != nil {
		t.Error("When calling a failed server, the client with error should NOT return an error")
	}
	v, ok := resp.(testStruct)
	if ok == false {
		t.Errorf("The returned response must be decoded, got %#v", resp)
	}
	if v.err == nil {
		t.Error("The response has not been correctly decoded")
	}

	if !IsRetryable(v.err) {
		t.Error("The returned error for http.StatusServiceUnavailable should be retryable")
	}

	h.statuses = []int{http.StatusBadRequest}
	resp, err = e(context.TODO(), nil)
	if err != nil {
		t.Error("When calling a failed server, the client with error should NOT return an error")
	}
	v, ok = resp.(testStruct)
	if ok == false {
		t.Errorf("The returned response must be decoded, got %#v", resp)
	}
	if v.err == nil {
		t.Error("The response has not been correctly decoded")
	}
	if IsRetryable(v.err) {
		t.Error("The returned error for http.StatusBadRequest should not be retryable")
	}


	h.statuses = []int{}
	res, err := e(context.TODO(), nil)
	if err != nil {
		t.Errorf("When calling a working server, the client should not return an error, got %s", err)
	} else if !reflect.DeepEqual(res, testData) {
		t.Errorf("The endpoint returned invalid data : %+v", res)
	}
}

var testData = testStruct{Foo: "bar"}

type testHandler struct {
	statuses []int
}

func (h *testHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	if len(h.statuses) > 0 {
		w.WriteHeader(h.statuses[0])
		h.statuses = h.statuses[1:]
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(testData)
}

func decodeTestResponse(ctx context.Context, resp *http.Response) (interface{}, error) {
	response := testStruct{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	response.err = HTTPError(resp)

	return response, nil
}
