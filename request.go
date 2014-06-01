package teaspoon

import (
	"bytes"
	"errors"
	"io"
)

type Request struct {
	Priority  byte
	Method    byte
	Resource  int16
	RequestID []byte
	Payload   []byte
}

var (
	InvalidPacketSequence = errors.New("Invalid sequence of packets provided")
	InvalidRequestId      = errors.New("Invalid request ID provided")
	RequestNotReady       = errors.New("The packets for the request ID are not ready yet")
)

func constructRequest(packets []*Packet, requestId []byte) (*Request, error) {
	if packets == nil {
		return nil, InvalidPacketSequence
	}

	if requestId == nil {
		return nil, InvalidRequestId
	}

	request_packets := make([]*Packet, 0)

	for i := range packets {
		if bytes.Compare(packets[i].requestId, requestId) == 0 {
			request_packets = append(request_packets, packets[i])

			if packets[i].sequence == packets[i].totalSequences-1 {
				return &Request{
					Priority:  packets[i].priority,
					Method:    packets[i].method,
					Resource:  packets[i].resource,
					RequestID: packets[i].requestId,
					Payload:   combinePacketPayloads(request_packets),
				}, nil
			}
		}
	}

	return nil, RequestNotReady
}

func ReadRequest(r io.Reader) (*Request, error) {
	packets := make([]*Packet, 0)

	for {
		packet, err := ReadPacket(r)
		if err != nil {
			return nil, err
		}

		packets = append(packets, packet)

		if packet.sequence == packet.totalSequences-1 {
			return constructRequest(packets, packet.requestId)
		}
	}

	return nil, io.EOF
}
