package binders

import (
	"io"
)

type ConnectionPool struct {
	rwcs []io.ReadWriteCloser
}

func (p *ConnectionPool) GetConnections() []io.ReadWriteCloser {
	return p.rwcs
}

func (p *ConnectionPool) OnClientConnect(rwc io.ReadWriteCloser) error {
	p.rwcs = append(p.rwcs, rwc)
	return nil
}

func (p *ConnectionPool) OnClientDisconnect(rwc io.ReadWriteCloser) {
	for i := 0; i < len(p.rwcs); i++ {
		if p.rwcs[i] == rwc {
			p.rwcs[i] = nil
			return
		}
	}
}
