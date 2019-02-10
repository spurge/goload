package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"gopkg.in/jarcoal/httpmock.v1"
)

type FakeParser struct {
	CallCounter int
}

func (f *FakeParser) Parse(input string) string {
	f.CallCounter += 1
	return input + "-faked"
}

func (f *FakeParser) Record(name, body string) {}

var faker HistoryHandler = &FakeParser{}

func TestLoadingRequests(t *testing.T) {
	content := []byte(`
- name: An request
  url: weird url
  method: Hunk
  body: '{"hested":"Ok"}'
  headers:
    key-: valz
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

	requests, err := LoadRequests(tmpfile.Name())

	if err != nil {
		t.Fatal(err)
	}

	r := requests[0]

	if r.GetName() != "An request" ||
		r.GetUrl() != "weird url" ||
		r.GetMethod() != "Hunk" ||
		r.GetBody() != `{"hested":"Ok"}` ||
		r.GetHeader("key-") != "valz" {
		t.Errorf("Request does not match: %#v", r)
	}

	err = tmpfile.Close()

	if err != nil {
		t.Fatal(err)
	}
}

func TestGetWith4xx(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	faker := FakeParser{}

	request := Request{
		URL:    "https://some-host",
		Method: "GET",
		Parser: &faker,
	}

	httpmock.RegisterResponder(
		request.Method,
		request.URL+"-faked",
		func(req *http.Request) (*http.Response, error) {
			return httpmock.NewStringResponse(404, "Hello!"), nil
		},
	)

	response, err := request.Send()

	if err != nil {
		t.Fatal("Send returned with an error")
	}

	if response.StatusCode != "4xx" || response.Body != "Hello!" {
		t.Errorf(
			"Response did not contain matching status code and/or body. %s != 4xx && %s != Hello!",
			response.StatusCode,
			response.Body,
		)
	}

	if faker.CallCounter != 1 {
		t.Error("Parse faker was not called once")
	}
}

func TestPostWithHeaders(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	faker := FakeParser{}

	request := Request{
		URL:     "https://some-host",
		Method:  "POST",
		Body:    "{\"property\":\"Some cool post body\"",
		Headers: map[string]string{"An Header": "With value"},
		Parser:  &faker,
	}

	httpmock.RegisterResponder(
		request.Method,
		request.URL+"-faked",
		func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("An Header-faked") != request.Headers["An Header"]+"-faked" {
				return httpmock.NewStringResponse(500, "Request without header"), nil
			}

			body, err := ioutil.ReadAll(req.Body)
			defer req.Body.Close()

			if err != nil || string(body) != request.GetBody()+"-faked" {
				return httpmock.NewStringResponse(500, "Request body does not match"), nil
			}

			return httpmock.NewStringResponse(200, "Hello!"), nil
		},
	)

	response, err := request.Send()

	if err != nil {
		t.Fatalf("Send returned with an error: %s", err)
	}

	if response.StatusCode != "2xx" || response.Body != "Hello!" {
		t.Errorf(
			"Response did not contain matching status code and/or body. %s != 2xx && %s != Hello!",
			response.StatusCode,
			response.Body,
		)
	}

	if faker.CallCounter != 4 {
		t.Error("Parse faker was not called four times")
	}
}

func TestGetName(t *testing.T) {
	r := Request{
		Name: "a name",
	}

	if r.GetName() != r.Name {
		t.Error("GetName() does not return Name")
	}
}

func TestSetParser(t *testing.T) {
	r := Request{}
	p := FakeParser{}

	r.SetParser(&p)

	if r.Parser != &p {
		t.Error("SetParser does not set correct parser")
	}
}

func TestIterator(t *testing.T) {
	c := RequestCollection{
		Requests: []*Request{
			&Request{Name: "1"},
			&Request{Name: "2"},
			&Request{Name: "3"},
			&Request{Name: "4"},
			&Request{Name: "5"},
			&Request{Name: "6"},
			&Request{Name: "7"},
		},
	}

	o := ""

	for i := 0; i < 3; i++ {
		for r := c.First(); r != nil; r = c.Next() {
			o += r.GetName()
		}
	}

	if o != "123456712345671234567" {
		t.Error("Iterator runs in wrong order")
	}
}
