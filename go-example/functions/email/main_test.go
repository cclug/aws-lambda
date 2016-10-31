package main

import "testing"

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
