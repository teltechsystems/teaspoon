package binders

import (
	"github.com/teltechsystems/teaspoon"
	"io"
	"math/rand"
	"time"
)

type Pinger struct {
	interval time.Duration
	rwcs     []io.ReadWriteCloser
	work     chan io.ReadWriteCloser
}

func (p *Pinger) OnClientConnect(rwc io.ReadWriteCloser) error {
	p.rwcs = append(p.rwcs, rwc)
	return nil
}

func (p *Pinger) OnClientDisconnect(rwc io.ReadWriteCloser) {
	for i := 0; i < len(p.rwcs); i++ {
		if p.rwcs[i] == rwc {
			p.rwcs[i] = nil
			return
		}
	}
}

func (p *Pinger) processPings() {
	n := int64(0)
	intervalInt := int64(p.interval / time.Second)

	for {
		for i := range p.rwcs {
			if int64(i)%intervalInt == 0 && p.rwcs[i] != nil {
				p.sendPing(p.rwcs[i])
			}
		}
		n = (n + 1) % intervalInt
		time.Sleep(p.interval)
	}
}

func (p *Pinger) sendPing(rwc io.ReadWriteCloser) {
	requestID := teaspoon.RequestID{}
	for i := 0; i < 16; i++ {
		requestID[i] = byte(rand.Intn(16))
	}

	r := &teaspoon.Request{
		OpCode:    teaspoon.OPCODE_PING,
		Priority:  5,
		Method:    0,
		Resource:  0,
		RequestID: requestID,
		Payload:   []byte{},
	}

	r.WriteTo(rwc)
}

func NewPinger(interval time.Duration) *Pinger {
	pinger := &Pinger{
		interval: interval,
		rwcs:     make([]io.ReadWriteCloser, 0),
		work:     make(chan io.ReadWriteCloser),
	}

	go pinger.processPings()

	return pinger
}
