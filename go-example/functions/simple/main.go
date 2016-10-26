package main

import (
	"encoding/json"
	"fmt"

	"github.com/apex/go-apex"
)

type message struct {
	Hello  string `json:"hello"`
	Ignore string `json:"ignore"`
}

func main() {
	apex.HandleFunc(func(event json.RawMessage, ctx *apex.Context) (interface{}, error) {
		return handle(event)
	})
}

func handle(event json.RawMessage) (string, error) {
	var view struct{ Hello string }

	if err := json.Unmarshal(event, &view); err != nil {
		return "", err
	}

	out := fmt.Sprintf("received greeting from: %s", view.Hello)
	return out, nil
}
