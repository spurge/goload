package main

import (
	"bytes"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
				logrus.
					WithField("function", "fromJson").
					WithField("entry", name).
					WithField("path", path).
					Error("Missing json template")
				return ""
			},
			"uuid": func() uuid.UUID {
				return uuid.New()
			},
			"now": func() time.Time {
				return time.Now()
			},
			"add": func(values ...int) int {
				add := 0

				for _, v := range values {
					add += v
				}

				return add
			},
			"sub": func(values ...int) int {
				if len(values) <= 0 {
					return 0
				}

				sub := values[0]

				for i := 1; i < len(values); i++ {
					sub -= values[i]
				}

				return sub
			},
			"mul": func(values ...int) int {
				if len(values) <= 0 {
					return 0
				}

				mul := values[0]

				for i := 1; i < len(values); i++ {
					mul *= values[i]
				}

				return mul
			},
		}).
		Parse(input)

	if err != nil {
		ParseTemplateError.Inc()
		logrus.
			WithError(err).
			Error("Error parsing templated input")
		return input
	}

	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, nil)

	if err != nil {
		ExecuteTemplateError.Inc()
		logrus.
			WithError(err).
			Error("Error executing template")
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
