package kitty

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
)

func newTestLogger(w io.Writer) log.Logger {
	return &testLogger{w}
}

type testLogger struct{ w io.Writer }

func (l *testLogger) Log(keyvals ...interface{}) error {
	for i := 0; i < len(keyvals); i++ {
		if keyvals[i] == "duration" {
			i++
			continue
		}
		io.WriteString(l.w, fmt.Sprintf("%v,", keyvals[i]))
	}
	return nil
}

func TestLogEndpoint(t *testing.T) {
	tcs := []struct {
		opt      LogOption
		response interface{}
		err      error
		log      string
	}{
		{
			opt:      LogRequest,
			response: "foo",
			log:      `msg,request: bar,`,
		},
		{
			opt:      LogResponse,
			response: "foo",
			log:      `status,200,msg,response: foo,`,
		},
		{
			opt:      LogErrors,
			response: "foo",
			log:      ``,
		},
		{
			opt: LogErrors,
			err: errors.New("bar"),
			log: `error,bar,status,500,msg,request: bar,`,
		},
	}
	for _, tc := range tcs {
		buf := bytes.NewBuffer([]byte{})
		ctx := context.WithValue(context.TODO(), logKey, newTestLogger(buf))
		e := func(_ context.Context, _ interface{}) (interface{}, error) {
			return tc.response, tc.err
		}
		LogEndpoint(tc.opt)(e)(ctx, "bar")
		logged := strings.TrimSpace(buf.String())
		if logged != tc.log {
			t.Errorf("Invalid log `%s` should have been `%s`", logged, tc.log)
		}
	}
}
