package binders

import (
	"io"
)

type ConnectionPool struct {
	writers []io.Writer
}

func (p *ConnectionPool) GetConnections() []io.Writer {
	return p.writers
}

func (p *ConnectionPool) OnClientConnect(w io.Writer) error {
	p.writers = append(p.writers, w)
	return nil
}

func (p *ConnectionPool) OnClientDisconnect(w io.Writer) {
	index := -1

	for i := 0; i < len(p.writers); i++ {
		if p.writers[i] == w {
			index = i
			break
		}
	}

	if index != -1 {
		p.writers = append(p.writers[0:index], p.writers[index+1:len(p.writers)]...)
	}
}
