package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/golang/glog"
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
	Params  map[string]string `yaml:"params"`
	Method  string            `yaml:"method"`
	Body    string            `yaml:"body"`
	Headers map[string]string `yaml:"headers"`
	Expect  Expected          `yaml:"expect"`
	Parser  HistoryHandler
}

func (r *Request) GetName() string {
	return r.Name
}

func (r *Request) GetUrl() string {
	url, err := url.Parse(r.Parser.Parse(r.URL))

	if err != nil {
		ParseURLError.Inc()
		return url.String()
	}

	query := url.Query()

	for k, v := range r.Params {
		query.Set(r.Parser.Parse(k), r.Parser.Parse(v))
	}

	url.RawQuery = query.Encode()

	return url.String()
}

func (r *Request) GetMethod() string {
	return r.Method
}

func (r *Request) GetHeader(key string) string {
	return r.Parser.Parse(r.Headers[key])
}

func (r *Request) GetHeaders() map[string]string {
	headers := make(map[string]string)

	for k := range r.Headers {
		headers[r.Parser.Parse(k)] = r.GetHeader(k)
	}

	return headers
}

func (r *Request) GetBody() string {
	return r.Parser.Parse(r.Body)
}

func (r *Request) SetParser(parser HistoryHandler) {
	r.Parser = parser
}

func (r *Request) Send() (Response, error) {
	var rec Response
	var req *http.Request
	var err error

	if r.Body != "" {
		req, err = http.NewRequest(
			r.GetMethod(),
			r.GetUrl(),
			bytes.NewBuffer([]byte(r.GetBody())),
		)
	} else {
		req, err = http.NewRequest(
			r.GetMethod(),
			r.GetUrl(),
			nil,
		)
	}

	if err != nil {
		return rec, err
	}

	for k, v := range r.GetHeaders() {
		req.Header.Set(k, v)
	}

	glog.Infof("Sending request to %s %s", req.Method, req.URL)

	then := time.Now()
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		RequestStatusCounter.WithLabelValues(r.GetName(), "error").Inc()
		glog.Errorf("Could not connect to %s %s: %s", req.Method, req.URL, err)
		return rec, err
	}

	bodybytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		RequestStatusCounter.WithLabelValues(r.GetName(), "error").Inc()
		glog.Errorf("Could not read body from %s %s: %s", req.Method, req.URL, err)
		return rec, err
	}

	latency := time.Since(then).Seconds()

	glog.Infof("Got response %d from %s %s", res.StatusCode, req.Method, req.URL)

	body := string(bodybytes)

	if res.StatusCode >= 400 {
		glog.Warningf("Response body from %s: %s", req.URL, body)
	}

	r.Expect.Evaluate(r.GetName(), res, body)

	rec = Response{Latency: latency, Body: body}
	rec.SetStatusCode(res.StatusCode)

	RequestStatusCounter.WithLabelValues(r.GetName(), rec.StatusCode).Inc()
	RequestLatencySummary.WithLabelValues(r.GetName(), rec.StatusCode).Observe(latency)

	return rec, nil
}

type Response struct {
	Latency        float64
	StatusCode     string
	RealStatusCode int
	Body           string
}

func (r *Response) SetStatusCode(statusCode int) {
	r.RealStatusCode = statusCode

	if statusCode < 400 {
		r.StatusCode = "2xx"
		return
	}

	if statusCode < 500 {
		r.StatusCode = "4xx"
		return
	}

	r.StatusCode = "5xx"
}
