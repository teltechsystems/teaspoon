package teaspoon

import (
	"errors"
	"io"
	"sync"
)

type Request struct {
	Priority  byte
	Method    byte
	Resource  int
	RequestID requestId
	Payload   []byte
}

type requestId [16]byte

var readerPackets map[io.Reader](map[requestId][]*Packet)
var readerPacketsMutex = sync.Mutex{}

var (
	InvalidPacketSequence = errors.New("Invalid sequence of packets provided")
	InvalidRequestId      = errors.New("Invalid request ID provided")
	RequestNotReady       = errors.New("The packets for the request ID are not ready yet")
)

func constructRequest(packets []*Packet) (*Request, error) {
	if packets == nil {
		return nil, InvalidPacketSequence
	}

	return &Request{
		Priority:  packets[0].priority,
		Method:    packets[0].method,
		Resource:  packets[0].resource,
		RequestID: packets[0].requestId,
		Payload:   combinePacketPayloads(packets),
	}, nil
}

func ReadRequest(r io.Reader) (*Request, error) {
	if readerPackets == nil {
		readerPackets = make(map[io.Reader]map[requestId][]*Packet)
	}

	for {
		packet, err := ReadPacket(r)
		if err != nil {
			return nil, err
		}

		readerPacketsMutex.Lock()
		if readerPackets[r] == nil {
			readerPackets[r] = make(map[requestId][]*Packet)
		}

		readerPackets[r][packet.requestId] = append(readerPackets[r][packet.requestId], packet)

		if packet.sequence == packet.totalSequences-1 {
			request, err := constructRequest(readerPackets[r][packet.requestId])
			delete(readerPackets[r], packet.requestId)
			readerPacketsMutex.Unlock()

			return request, err
		}

		readerPacketsMutex.Unlock()
	}

	return nil, io.EOF
}
