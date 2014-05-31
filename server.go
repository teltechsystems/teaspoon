package teaspoon

import (
	"io"
	"net"
)

type Server struct {
}

func (s *Server) Serve(l net.Listener) error {
	defer l.Close()

	for {
		rwc, err := l.Accept()
		if err != nil {
			return err
		}

		// No functionality yet
		rwc.Close()
	}

	return nil
}

type conn struct {
	rwc io.ReadWriteCloser
}

func (c *conn) serve() {
}
