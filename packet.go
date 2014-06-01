package teaspoon

import (
	"io"
)

type Packet struct {
	opCode         byte
	priority       byte
	method         byte
	resource       int16
	sequence       int32
	totalSequences int32
	requestId      []byte
	payloadLength  int32
	payload        []byte
}

func combinePacketPayloads(packets []*Packet) []byte {
	payloadLength := int32(0)
	for i := range packets {
		payloadLength += packets[i].payloadLength
	}

	payload := make([]byte, payloadLength)

	offset := int32(0)

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
		if err != nil {
			return nil, err
		}

		total_bytes_read += bytes_read
	}

	packet.opCode = (header[0] & 0xF0) >> 4
	packet.priority = header[0] & 0x0F
	packet.method = header[1] & 0x0F
	packet.resource = (int16(header[2]) << 8) + int16(header[3])
	packet.sequence = (int32(header[4]) << 8) + int32(header[5])
	packet.totalSequences = (int32(header[6]) << 8) + int32(header[7])
	packet.requestId = header[8:24]
	packet.payloadLength = (int32(header[24]) << 24) + (int32(header[25]) << 16) +
		(int32(header[26]) << 8) + int32(header[27])

	packet.payload = make([]byte, packet.payloadLength)

	for total_bytes_read := int32(0); total_bytes_read < packet.payloadLength; {
		bytes_read, err := r.Read(packet.payload)
		if err != nil {
			return nil, err
		}

		total_bytes_read += int32(bytes_read)
	}

	return packet, nil
}
