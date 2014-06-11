Teaspoon (TSP)
=========

Teaspoon is an internal protocol for TelTech services. This particular implementation is created in Go to provide a simple backbone for future applications written using Teaspoon.

Many of the design principles integrated into the package were followed by Go's standard HTTP package. You'll notice a similar implementation in many areas, with the advantage of asynchronous communication and a single multiplexed connection to handle multiple requests on a single socket simultaneously.

A Simple Echo Server
--------------------
```go
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
```

License
----

MIT
