package main

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"gopkg.in/yaml.v2"
)

func LoadRequests(filename string) ([]*Request, error) {
	data, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	var requests []*Request

	err = yaml.Unmarshal(data, &requests)

	return requests, err
}

type RequestCollectionHandler interface {
	Len() int
	First() RequestHandler
	Next() RequestHandler
}

var requestCollection RequestCollectionHandler = &RequestCollection{}

type RequestCollection struct {
	Index    int
	Requests []*Request
}

func (r *RequestCollection) Len() int {
	return len(r.Requests)
}

func (r *RequestCollection) First() RequestHandler {
	r.Index = 0
	return r.Current()
}

func (r *RequestCollection) Next() RequestHandler {
	if r.Index < len(r.Requests) {
		r.Index++
	}

	return r.Current()
}

func (r *RequestCollection) Current() RequestHandler {
	if r.Index < len(r.Requests) {
		return r.Requests[r.Index]
	}

	return nil
}

type RequestHandler interface {
	SetParser(HistoryHandler)
	GetName() string
	GetUrl() string
	GetMethod() string
	GetBody() string
	GetHeader(key string) string
	Send() (Response, error)
}

var requestHandler RequestHandler = &Request{}

type Request struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Method  string            `yaml:"method"`
	Body    string            `yaml:"body"`
	Headers map[string]string `yaml:"headers"`
	Parser  HistoryHandler
}

func (r *Request) GetName() string {
	return r.Name
}

func (r *Request) GetUrl() string {
	return r.URL
}

func (r *Request) GetMethod() string {
	return r.Method
}

func (r *Request) GetHeader(key string) string {
	return r.Headers[key]
}

func (r *Request) GetBody() string {
	return r.Body
}

func (r *Request) SetParser(parser HistoryHandler) {
	r.Parser = parser
}

func (r *Request) Send() (Response, error) {
	var response Response
	var req *http.Request
	var err error

	if r.Body != "" {
		req, err = http.NewRequest(
			r.Method,
			r.Parser.Parse(r.URL),
			bytes.NewBuffer([]byte(r.Parser.Parse(r.GetBody()))),
		)
	} else {
		req, err = http.NewRequest(
			r.Method,
			r.Parser.Parse(r.URL),
			nil,
		)
	}

	if err != nil {
		return response, err
	}

	for key, value := range r.Headers {
		req.Header.Set(r.Parser.Parse(key), r.Parser.Parse(value))
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return response, err
	}

	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return response, err
	}

	response = Response{Body: string(body)}
	response.SetStatusCode(res.StatusCode)

	return response, nil
}

type Response struct {
	StatusCode string
	Body       string
}

func (r *Response) SetStatusCode(statusCode int) {
	if statusCode < 200 {
		r.StatusCode = "1xx"
		return
	}

	if statusCode < 300 {
		r.StatusCode = "2xx"
		return
	}

	if statusCode < 400 {
		r.StatusCode = "3xx"
		return
	}

	if statusCode < 500 {
		r.StatusCode = "4xx"
		return
	}

	r.StatusCode = "5xx"
}
