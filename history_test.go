package main

import (
	"regexp"
	"testing"
)

func TestAddAndParseJsonRecord(t *testing.T) {
	json := `{"a":{"property":[{"value":"YEZ"}]}}`
	input := `{{ fromJson "some name" "a.property.0.value" }}`

	history := NewHistory()
	history.Record("some name", json)

	output := history.Parse(input)

	if output != "YEZ" {
		t.Errorf("Parser did not render template into YEZ, instead: %s", output)
	}
}

func TestMissingRecord(t *testing.T) {
	input := `{{ fromJson "some name" "a.property.0.value" }}`

	history := NewHistory()
	output := history.Parse(input)

	if output != "" {
		t.Errorf("Parser did not render template empty: %s", output)
	}
}

func TestUuidTemplateFuncs(t *testing.T) {
	input := `{{ (uuid).String }}`

	history := NewHistory()
	output := history.Parse(input)

	re := regexp.MustCompile("^[a-z0-9]{8}-([a-z0-9]{4}-){3}[a-z0-9]{12}$")

	if !re.MatchString(output) {
		t.Errorf("Now didn't match, returned %s", output)
	}
}

func TestNowTemplateFuncs(t *testing.T) {
	input := `{{ (now.Add 123).Unix }}`

	history := NewHistory()
	output := history.Parse(input)

	re := regexp.MustCompile("^[0-9]{10,}$")

	if !re.MatchString(output) {
		t.Errorf("Now didn't match, returned %s", output)
	}
}

func TestAddTemplateFuncs(t *testing.T) {
	input := `{{ add 3 8 21 }}`

	history := NewHistory()
	output := history.Parse(input)

	if output != "32" {
		t.Errorf("Add didn't match, returned %s", output)
	}
}

func TestSubTemplateFuncs(t *testing.T) {
	input := `{{ sub 21 9 4 1 }}`

	history := NewHistory()
	output := history.Parse(input)

	if output != "7" {
		t.Errorf("Sub didn't match, returned %s", output)
	}
}

func TestMulTemplateFuncs(t *testing.T) {
	input := `{{ mul 4 1 9 }}`

	history := NewHistory()
	output := history.Parse(input)

	if output != "36" {
		t.Errorf("Mul didn't match, returned %s", output)
	}
}
