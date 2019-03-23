package main

import (
	"net/http"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestEvaluate(t *testing.T) {
	content := []byte(`
status_code_re: 200
headers_re:
  Content-Type: application.*
body_re: .*`)

	var e Expected

	err := yaml.Unmarshal(content, &e)

	if err != nil {
		t.Fatal(err)
	}

	r := http.Response{
		StatusCode: 200,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}
	b := ""

	err = e.Evaluate("some name", &r, b)

	if e.Name != "some name" {
		t.Error("Name was not set")
	}

	if err != nil {
		t.Errorf("Should not return error on success: %s", err)
	}
}

func TestEvaluateStatusCode(t *testing.T) {
	e := Expected{
		StatusCode: "([0-9])+",
	}

	err := e.EvaluateStatusCode(200)

	if err != nil {
		t.Errorf("Should not return error on success: %s", err)
	}
}

func TestEvaluateStatusCodeMissmatch(t *testing.T) {
	e := Expected{
		StatusCode: "3[0-9]{2}",
	}

	err := e.EvaluateStatusCode(200)

	if err == nil {
		t.Error("Should not return nil error on failure")
	}
}

func TestEvaluateStatusCodeCompileError(t *testing.T) {
	e := Expected{
		StatusCode: "3[}",
	}

	err := e.EvaluateStatusCode(200)

	if err == nil {
		t.Error("Should not return nil error on failure")
	}
}

func TestEvaluateHeaders(t *testing.T) {
	e := Expected{
		Headers: map[string]string{
			"Some-Header": "^[a-z]+$",
		},
	}

	h := http.Header{
		"Some-Header": []string{"abc"},
	}

	err := e.EvaluateHeaders(&h)

	if err != nil {
		t.Errorf("Should not return error on success: %s", err)
	}
}

func TestEvaluateHeadersMissmatch(t *testing.T) {
	e := Expected{
		Headers: map[string]string{
			"Some-Header": "^[a-z]+$",
		},
	}

	h := http.Header{
		"Some-Header": []string{"123"},
	}

	err := e.EvaluateHeaders(&h)

	if err == nil {
		t.Errorf("Should not return nil error on failure")
	}
}

func TestEvaluateHeadersCompileError(t *testing.T) {
	e := Expected{
		Headers: map[string]string{
			"Some-Header": "^}[a-z]+$",
		},
	}

	h := http.Header{
		"Some-Header": []string{"123"},
	}

	err := e.EvaluateHeaders(&h)

	if err == nil {
		t.Errorf("Should not return nil error on failure")
	}
}

func TestEvaluateBody(t *testing.T) {
	e := Expected{
		Body: "a{3}",
	}

	err := e.EvaluateBody("hisdhaaa93")

	if err != nil {
		t.Errorf("Should not return error on success: %s", err)
	}
}

func TestEvaluateBodyMissmatch(t *testing.T) {
	e := Expected{
		Body: "[0-9]+",
	}

	err := e.EvaluateBody("abc")

	if err == nil {
		t.Errorf("Should not return nil error on failure")
	}
}

func TestEvaluateBodyCompileError(t *testing.T) {
	e := Expected{
		Body: "[0-](/+",
	}

	err := e.EvaluateBody("abc")

	if err == nil {
		t.Errorf("Should not return nil error on failure")
	}
}
