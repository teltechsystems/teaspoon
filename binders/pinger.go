package binders

import (
	"github.com/teltechsystems/teaspoon"
	"io"
	"math/rand"
	"time"
)

type Pinger struct {
	ConnectionPool
	interval time.Duration
}

func (p *Pinger) processPings() {
	intervalInt := int64(p.interval / time.Second)

	for {
		for i := 0; i < int(intervalInt); i++ {
			connections := p.GetConnections()
			for j := i; j < len(connections); j += int(intervalInt) {
				if p.rwcs[j] != nil {
					p.sendPing(connections[j])
				}
			}
			time.Sleep(time.Second)
		}
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
	pinger := &Pinger{interval: interval}

	go pinger.processPings()

	return pinger
}
