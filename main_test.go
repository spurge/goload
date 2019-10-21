package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"gopkg.in/jarcoal/httpmock.v1"
)

func TestRequests(t *testing.T) {
	content := []byte(`
- name: request-1
  url: http://some-url-1
  params:
    param: value
  method: GET
  expect:
    status_code_re: '2..'
    body_re: '"name":\s*"edud"'
    headers_re:
      Content-Length: '[0-9]+'
- name: request-2
  url: 'http://some-url-2/{{ fromJson "request-1" "auth.path" }}?param=value'
  method: POST
  headers:
    Authorization: '{{ fromJson "request-1" "auth.token" }}'
  body: '{
      "username": "hello:{{ fromJson "request-1" "auth.name" }}",
      "password": "secret"
    }'
- name: request-3
  url: 'http://some-url-3/{{ fromJson "request-1" "auth.path" }}/ok'
  method: PUT
  headers:
    If-Match: '{{ fromJson "request-2" "user.checksum" }}'
  body: '{
      "user": {
        "firstname": "dude"
      }
    }'
`)

	tmpfile, err := ioutil.TempFile("", "*")

	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.Write(content)

	if err != nil {
		t.Fatal(err)
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	wait := make(chan int, 6)

	httpmock.RegisterResponder(
		"GET",
		"http://some-url-1",
		func(req *http.Request) (*http.Response, error) {
			wait <- 1

			return httpmock.NewStringResponse(200, `{
				"auth": {
					"path": "a-path",
					"name": "edud",
					"token": "crazy token"
				}
			}`), nil
		},
	)

	httpmock.RegisterResponder(
		"POST",
		"http://some-url-2/a-path",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("Authorization") != "crazy token" {
				t.Error("Request two did not contain crazy token")
			}

			body, err := ioutil.ReadAll(req.Body)
			defer req.Body.Close()

			if err != nil {
				t.Error(err)
			}

			if string(body) != `{ "username": "hello:edud", "password": "secret" }` {
				t.Errorf("Request two body did not match: %s", string(body))
			}

			wait <- 2

			return httpmock.NewStringResponse(200, `{
				"user": {
					"checksum": "md5"
				}
			}`), nil
		},
	)

	httpmock.RegisterResponder(
		"PUT",
		"http://some-url-3/a-path/ok",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("If-Match") != "md5" {
				t.Error("Request three did not contain checksum")
			}

			body, err := ioutil.ReadAll(req.Body)
			defer req.Body.Close()

			if err != nil {
				t.Error(err)
			}

			if string(body) != `{ "user": { "firstname": "dude" } }` {
				t.Errorf("Request three body did not match: %s", string(body))
			}

			wait <- 3

			return httpmock.NewStringResponse(200, ""), nil
		},
	)

	status := NewStatus()
	go InitiateRequests(2, 1, -1, tmpfile.Name(), status, make(chan bool))
	go func() {
		time.Sleep(4 * time.Second)
		t.Error("Timeout")
		close(wait)
	}()

	jobs := map[int]int{1: 0, 2: 0, 3: 0}

	for w := range wait {
		jobs[w]++

		if jobs[1] >= 4 && jobs[2] >= 4 && jobs[3] >= 4 {
			close(wait)
		}
	}
}

func TestLimitedRepeat(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	called := 0
	wait := make(chan bool)

	httpmock.RegisterResponder(
		"GET",
		"http://some-url-1",
		func(req *http.Request) (*http.Response, error) {
			called++

			return httpmock.NewStringResponse(200, ""), nil
		},
	)

	requests := []*Request{
		&Request{
			URL:    "http://some-url-1",
			Method: "GET",
		},
	}

	go func() {
		time.Sleep(4 * time.Second)
		t.Error("Timeout")
		close(wait)
	}()

	status := NewStatus()
	go RunRequests(requests, 0, 2, status, wait)

	<-wait

	if called != 3 {
		t.Errorf("Repeated %d times instead of 3", called)
	}
}
