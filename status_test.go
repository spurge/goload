package main

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

type ResponseMock struct {
	WriteCallback func([]byte) (int, error)
}

func (r *ResponseMock) Header() http.Header {
	return http.Header{}
}

func (r *ResponseMock) WriteHeader(status int) {}

func (r *ResponseMock) Write(body []byte) (int, error) {
	return r.WriteCallback(body)
}

func TestRecordAndServe(t *testing.T) {
	called := 0
	fixture := `{"slowest":{"a request":[{"latency":123.3,"status":200,"response":{"ok":"yes?"}},{"latency":1.123,"status":200,"response":{"ok":"yes?"}}],"another request":[{"latency":333,"status":200,"response":{"ok":"yes?"}},{"latency":93.1,"status":200,"response":{"ok":"yes?"}},{"latency":12.3,"status":200,"response":{"ok":"yes?"}}]},"errors":{"a request":[],"another request":[{"latency":12.3,"status":200,"response":{"ok":"yes?"},"error":"some error"},{"latency":12.3,"status":200,"response":{"ok":"yes?"},"error":"some other error"}]}}`
	res := ResponseMock{
		WriteCallback: func(body []byte) (int, error) {
			called++

			if string(body) != fixture {
				t.Errorf("Body didn't match\n%s", string(body))
			}

			return 0, nil
		},
	}

	status := NewStatus()

	status.Record("a request", 1.123, 200, `{"ok":"yes?"}`, nil)
	status.Record("a request", 123.3, 200, `{"ok":"yes?"}`, nil)
	status.Record("another request", 12.3, 200, `{"ok":"yes?"}`, nil)
	status.Record("another request", 93.1, 200, `{"ok":"yes?"}`, nil)
	status.Record("another request", 12.3, 200, `{"ok":"yes?"}`, nil)
	status.Record("another request", 333.0, 200, `{"ok":"yes?"}`, nil)
	status.Record("another request", 12.3, 200, `{"ok":"yes?"}`, errors.New("error gone"))
	status.Record("another request", 12.3, 200, `{"ok":"yes?"}`, errors.New("an error"))
	status.Record("another request", 12.3, 200, `{"ok":"yes?"}`, errors.New("some other error"))
	status.Record("another request", 12.3, 200, `{"ok":"yes?"}`, errors.New("some error"))

	time.Sleep(time.Millisecond * 100)

	if len(status.Responses) > 0 {
		t.Error("Record responses took too long, more than 100 ms")
	}

	status.ServeHTTP(&res, &http.Request{})

	if called != 1 {
		t.Error("Response writer Write wasn't called once")
	}
}
