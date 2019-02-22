package main

import (
	"bytes"
	"text/template"
	"time"

	"github.com/golang/glog"
	"github.com/tidwall/gjson"
)

type HistoryHandler interface {
	Record(name, body string)
	Parse(input string) string
}

var historyHandler HistoryHandler = &History{}

type History struct {
	Records map[string]*Record
}

func NewHistory() *History {
	return &History{
		Records: make(map[string]*Record),
	}
}

func (h *History) Record(name, result string) {
	h.Records[name] = &Record{
		Body: result,
	}
}

func (h *History) Parse(input string) string {
	tmpl, err := template.
		New("History parser").
		Funcs(template.FuncMap{
			"fromJson": func(name, path string) string {
				r := h.From(name)

				if r != nil {
					return r.Json(path)
				}

				MissingTemplateEntryError.Inc()
				glog.Errorf("Missing json template entry from %s with %s", name, path)
				return ""
			},
			"now": func() time.Time {
				return time.Now()
			},
		}).
		Parse(input)

	if err != nil {
		ParseTemplateError.Inc()
		glog.Errorf("Error parsing templated input: %s", err)
		return input
	}

	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, nil)

	if err != nil {
		ExecuteTemplateError.Inc()
		glog.Errorf("Error executing template: %s", err)
		return input
	}

	return buf.String()
}

func (h *History) From(name string) *Record {
	return h.Records[name]
}

type Record struct {
	Body string
}

func (r *Record) Json(path string) string {
	return gjson.Get(r.Body, path).String()
}
