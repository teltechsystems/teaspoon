package teaspoon

import (
	"errors"
	"io"
)

var (
	PacketPayloadLengthExceeded = errors.New("The payload's packet length is too large")
)

type Packet struct {
	opCode         byte
	priority       byte
	method         byte
	resource       int
	sequence       int32
	totalSequences int32
	requestId      RequestID
	payloadLength  uint32
	payload        []byte
}

func combinePacketPayloads(packets []*Packet) []byte {
	payloadLength := uint32(0)
	for i := range packets {
		payloadLength += packets[i].payloadLength
	}

	payload := make([]byte, payloadLength)

	offset := uint32(0)

	for i := range packets {
		copy(payload[offset:offset+packets[i].payloadLength], packets[i].payload)

		offset += packets[i].payloadLength
	}

	return payload
}

func ReadPacket(r io.Reader) (*Packet, error) {
	packet := new(Packet)
	header := make([]byte, 28)

	for total_bytes_read := 0; total_bytes_read < 28; {
		bytes_read, err := r.Read(header)
		// logger.Printf("ReadPacket - bytes read: %d", bytes_read)
		if err != nil {
			return nil, err
		}

		total_bytes_read += bytes_read
	}

	packet.opCode = (header[0] & 0xF0) >> 4
	packet.priority = header[0] & 0x0F
	packet.method = header[1] & 0x0F
	packet.resource = (int(header[2]) << 8) + int(header[3])
	packet.sequence = (int32(header[4]) << 8) + int32(header[5])
	packet.totalSequences = (int32(header[6]) << 8) + int32(header[7])
	copy(packet.requestId[:], header[8:24])
	packet.payloadLength = (uint32(header[24]) << 24) + (uint32(header[25]) << 16) +
		(uint32(header[26]) << 8) + uint32(header[27])

	// logger.Printf("ReadPacket - payloadLength: %d", packet.payloadLength)

	if packet.payloadLength > 1200 {
		return nil, PacketPayloadLengthExceeded
	}

	packet.payload = make([]byte, packet.payloadLength)
	for total_bytes_read := uint32(0); total_bytes_read < packet.payloadLength; {
		bytes_read, err := r.Read(packet.payload[total_bytes_read:])
		if err != nil {
			return nil, err
		}

		total_bytes_read += uint32(bytes_read)
	}

	// logger.Printf("ReadPacket - generated packet: %v", packet)

	return packet, nil
}
