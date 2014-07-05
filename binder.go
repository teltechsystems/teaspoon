package teaspoon

import (
	"io"
)

type Binder interface {
	OnClientConnect(c io.ReadWriteCloser) error
	OnClientDisconnect(c io.ReadWriteCloser)
}
