package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
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

// func TestEventToMail(t *testing.T) {
// 	event, err := ioutil.ReadFile("event_test.json")
// 	if err != nil {
// 		t.Error(err.Error())
// 	}
// 	m, _ := eventToMail(json.RawMessage(event))
// 	// if err != nil {
// 	// 	t.Errorf("%s", err.Error()) //
// 	// 	return
// 	// }
// 	var want = struct {
// 		from    string
// 		subject string
// 	}{
// 		"John Smith <john@mail.com>",
// 		"test subject",
// 	}

// 	if got := m.headers.from[0]; got != want.from {
// 		t.Errorf("from = %s, want: %s\n", got, want.from)
// 	}

// 	if got := m.headers.subject; got != want.subject {
// 		t.Errorf("subject = %s, want: %s\n", got, want.subject)
// 	}

// }

func TestGetText(t *testing.T) {
	body, err := ioutil.ReadFile("../../body2.txt")
	if err != nil {
		t.Error(err.Error())
	}
	msg, err := getText(body) //
	if err != nil {
		t.Errorf(err.Error()) //
	}
	breakLine(os.Stderr)
	fmt.Fprintf(os.Stderr, "body: %s\n", msg)
	breakLine(os.Stderr)
}

func TestPayload(t *testing.T) {
	from := "me@mail.com"
	text := "the email\n"
	messageId := "long message ID"
	subject := "the subject"

	buf := payload(from, text, subject, messageId)
	breakLine(os.Stderr)
	fmt.Fprintf(os.Stderr, "payload: \n\n%s\n", string(buf))
	breakLine(os.Stderr)
}

func breakLine(w io.Writer) {
	fmt.Fprintf(w, strings.Repeat("-", 20))
	fmt.Fprintf(w, "\n")
}
