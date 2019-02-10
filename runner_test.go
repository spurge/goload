package main

import (
	"fmt"
	"testing"
)

type RequestCollectionFaker struct {
	Index    int
	Requests []*RequestFaker
}

func (r *RequestCollectionFaker) Len() int {
	return len(r.Requests)
}

func (r *RequestCollectionFaker) First() RequestHandler {
	r.Index = 0
	return r.Requests[0]
}

func (r *RequestCollectionFaker) Next() RequestHandler {
	r.Index++

	if r.Index < r.Len() {
		return r.Requests[r.Index]
	}

	return nil
}

var requestCollectionFaker RequestCollectionHandler = &RequestCollectionFaker{}

type RequestFaker struct {
	Parser HistoryHandler
	Name   string
	Body   string
}

func (r *RequestFaker) SetParser(parser HistoryHandler) {
	r.Parser = parser
}

func (r *RequestFaker) GetName() string {
	return r.Name
}

func (r *RequestFaker) GetUrl() string {
	return ""
}

func (r *RequestFaker) GetMethod() string {
	return ""
}

func (r *RequestFaker) GetHeader(key string) string {
	return ""
}

func (r *RequestFaker) GetBody() string {
	return r.Body
}

func (r *RequestFaker) Send() (Response, error) {
	return Response{
		StatusCode: "2xx",
		Body:       fmt.Sprintf("response %s %s", r.Name, r.Body),
	}, nil
}

var requestFaker RequestHandler = &RequestFaker{}

type HistoryFaker struct {
	RecordCalls map[string]string
}

func (h *HistoryFaker) Record(name, body string) {
	h.RecordCalls[name] = body
}

func (h *HistoryFaker) Parse(input string) string {
	return ""
}

var historyFaker HistoryHandler = &HistoryFaker{}

func TestRun(t *testing.T) {
	requests := RequestCollectionFaker{Requests: []*RequestFaker{
		&RequestFaker{Name: "name 1", Body: "body 1"},
		&RequestFaker{Name: "name 2", Body: "body 2"},
	}}
	history := HistoryFaker{
		RecordCalls: make(map[string]string),
	}
	runner := Runner{
		Requests: &requests,
		History:  &history,
	}

	runner.Run()

	if history.RecordCalls["name 1"] != "response name 1 body 1" ||
		history.RecordCalls["name 2"] != "response name 2 body 2" {
		t.Error("Request send and history does not match")
	}
}
