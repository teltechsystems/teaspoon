package teaspoon

import (
	"errors"
	"io"
	"sync"
)

type RequestID [16]byte

type Request struct {
	OpCode    byte
	Priority  byte
	Method    byte
	Resource  int
	RequestID RequestID
	Payload   []byte
}

func (r *Request) GetFrames(frameSize int32) [][]byte {
	payload := r.Payload
	totalSequences := int32(len(payload))/frameSize + 1

	frames := [][]byte{}

	for sequence := int32(0); sequence < totalSequences; sequence++ {
		payloadLength := frameSize
		if int32(len(payload)) < (sequence+1)*frameSize {
			payloadLength = int32(len(payload)) - sequence*frameSize
		}

		frame := []byte{
			(r.OpCode << 4) | r.Priority, r.Method, byte(r.Resource >> 8), byte(r.Resource),
			byte(sequence >> 8), byte(sequence), byte(totalSequences >> 8), byte(totalSequences),
		}
		frame = append(frame, r.RequestID[:]...)
		frame = append(frame, []byte{byte(payloadLength >> 24), byte(payloadLength >> 16), byte(payloadLength >> 8), byte(payloadLength)}...)
		frame = append(frame, payload[sequence*frameSize:sequence*frameSize+payloadLength]...)

		frames = append(frames, frame)
	}

	return frames
}

func (r *Request) WriteTo(w io.Writer) (n int64, err error) {
	for _, frame := range r.GetFrames(1200) {
		bw, err := w.Write(frame)
		n += int64(bw)
		if err != nil {
			return n, err
		}
	}

	return n, err
}

var readerPackets map[io.Reader](map[RequestID][]*Packet)
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

	if len(packets) <= 0 {
		return nil, InvalidPacketSequence
	}

	return &Request{
		OpCode:    packets[0].opCode,
		Priority:  packets[0].priority,
		Method:    packets[0].method,
		Resource:  packets[0].resource,
		RequestID: packets[0].requestId,
		Payload:   combinePacketPayloads(packets),
	}, nil
}

func ReadRequest(r io.Reader) (*Request, error) {
	readerPacketsMutex.Lock()
	if readerPackets == nil {
		readerPackets = make(map[io.Reader]map[RequestID][]*Packet)
	}
	readerPacketsMutex.Unlock()

	for {
		packet, err := ReadPacket(r)
		if err != nil {
			return nil, err
		}

		readerPacketsMutex.Lock()
		if readerPackets[r] == nil {
			readerPackets[r] = make(map[RequestID][]*Packet)
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
