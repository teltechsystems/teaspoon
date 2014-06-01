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

		conn := &conn{rwc: rwc}
		go conn.serve()
	}

	return nil
}

type conn struct {
	rwc io.ReadWriteCloser
}

func (c *conn) serve() {
	defer c.rwc.Close()
}
