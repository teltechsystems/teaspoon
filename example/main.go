package main

import (
	"fmt"
	"github.com/teltechsystems/teaspoon"
)

func EchoHandler(w teaspoon.ResponseWriter, r *teaspoon.Request) {
	fmt.Fprintf(w, "ECHO %s", string(r.Payload))
}

func main() {
	teaspoon.ListenAndServe(":8000", teaspoon.HandlerFunc(EchoHandler))
}
