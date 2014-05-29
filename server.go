package teaspoon

import (
	// "fmt"
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
}
