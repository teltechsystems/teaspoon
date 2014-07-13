package teaspoon

import (
	"io"
)

type Binder interface {
	OnClientConnect(c io.Writer) error
	OnClientDisconnect(c io.Writer)
}
