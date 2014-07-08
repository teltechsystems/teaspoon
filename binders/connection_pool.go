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
	index := -1

	for i := 0; i < len(p.rwcs); i++ {
		if p.rwcs[i] == rwc {
			index = i
			break
		}
	}

	if index != -1 {
		p.rwcs = append(p.rwcs[0:index], p.rwcs[index+1:len(p.rwcs)]...)
	}
}
