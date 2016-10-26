package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHandle(t *testing.T) {
	var m message
	m.Hello = "test"
	m.Ignore = "ignore this"

	raw, err := json.Marshal(m)
	if err != nil {
		t.Errorf("json.Marshal returned: %s", err)
	}

	s, err := handle(raw)

	if err != nil {
		t.Errorf("handle returned: %s", err)
	}

	want := "test"
	if !strings.Contains(s, want) {
		t.Errorf("handle() = %s want: %s\n", s, want)
	}
}
