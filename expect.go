package main

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/sirupsen/logrus"
)

type Expected struct {
	Name       string
	StatusCode string            `yaml:"status_code_re"`
	Headers    map[string]string `yaml:"headers_re"`
	Body       string            `yaml:"body_re"`
}

func (e *Expected) Evaluate(name string, r *http.Response, b string) error {
	e.Name = name

	e.EvaluateStatusCode(r.StatusCode)
	e.EvaluateHeaders(&r.Header)
	e.EvaluateBody(b)

	return nil
}

func (e *Expected) EvaluateStatusCode(s int) error {
	counter := ExpectedResponseCounter.WithLabelValues(e.Name, "status_code")

	if e.StatusCode == "" {
		return nil
	}

	if match(e.StatusCode, fmt.Sprintf("%d", s)) {
		counter.Inc()
		return nil
	}

	return fmt.Errorf("Status code %d, did not match %s", s, e.StatusCode)
}

func (e *Expected) EvaluateHeaders(h *http.Header) error {
	counter := ExpectedResponseCounter.WithLabelValues(e.Name, "headers")
	errors := 0

	if len(e.Headers) == 0 {
		return nil
	}

	for k, v := range e.Headers {
		if match(v, h.Get(k)) {
			counter.Inc()
		} else {
			errors++
		}
	}

	if errors > 0 {
		return fmt.Errorf("Headers did not match")
	}

	return nil
}

func (e *Expected) EvaluateBody(b string) error {
	counter := ExpectedResponseCounter.WithLabelValues(e.Name, "body")

	if e.Body == "" {
		return nil
	}

	if match(e.Body, b) {
		counter.Inc()
		return nil
	}

	return fmt.Errorf("Body did not match %s", e.Body)
}

func match(exp, target string) bool {
	re, err := regexp.Compile(exp)

	if err != nil {
		ExpectReCompileError.Inc()
		logrus.
			WithError(err).
			WithField("regexp", exp).
			Error("Could not compile regular expression for expected evaluation")
		return false
	}

	return re.MatchString(target)
}
