package main

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestIsAuthSender(t *testing.T) {
	var tests = []struct {
		addr string
		want bool
	}{
		{"me@tobin.cc", true},
		{"ME@TOBIN.CC", true},
		{"Tobin Harding <me@tobin.cc>", true},

		{"Tobin Harding <invalid@tobin.cc>", false},
		{"invalid@mail.com", false},
	}

	for _, test := range tests {
		if got := isAuthSender(test.addr); got != test.want {
			t.Errorf("isAuth(%s)\n", test.addr)
		}
	}
}

// func TestWhitelistPtrs(t *testing.T) {
// 	ptrs := whitelistPtrs()
// 	for _, ptr := range ptrs {
// 		fmt.Fprintf(os.Stderr, "%s\n", *ptr)
// 	}
// }

func TestEventToMail(t *testing.T) {
	event, err := ioutil.ReadFile("event_test.json")
	if err != nil {
		t.Error(err.Error())
	}
	m, err := eventToMail(json.RawMessage(event))
	if err != nil {
		t.Error(err.Error())
		return
	}
	var want = struct {
		from    string
		subject string
	}{
		"John Smith <john@mail.com>",
		"test subject",
	}

	if got := m.headers.from[0]; got != want.from {
		t.Errorf("from = %s, want: %s\n", got, want.from)
	}

	if got := m.headers.subject; got != want.subject {
		t.Errorf("subject = %s, want: %s\n", got, want.subject)
	}

}
