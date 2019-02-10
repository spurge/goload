package main

import (
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
