package teaspoon

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
)

func BenchmarkCombinePackets(b *testing.B) {
	packets := []*Packet{
		&Packet{payloadLength: 1, payload: []byte{'a'}},
		&Packet{payloadLength: 1, payload: []byte{'b'}},
	}

	for i := 0; i < b.N; i++ {
		combinePacketPayloads(packets)
	}
}

func TestCombinePacketPayloads(t *testing.T) {
	Convey("An empty packet array should return an empty payload", t, func() {
		payload := combinePacketPayloads([]*Packet{})
		So(payload, ShouldResemble, []byte{})
	})

	Convey("A single packet in a slice should return an appropriate payload", t, func() {
		payload := combinePacketPayloads([]*Packet{
			&Packet{payloadLength: 1, payload: []byte{1}},
		})

		So(payload, ShouldResemble, []byte{1})
	})

	Convey("Multiple packets in a slice should return an appropriate payload", t, func() {
		payload := combinePacketPayloads([]*Packet{
			&Packet{payloadLength: 1, payload: []byte{1}},
			&Packet{payloadLength: 2, payload: []byte{2, 3}},
		})

		So(payload, ShouldResemble, []byte{1, 2, 3})
	})
}

func BenchmarkReadPacket(b *testing.B) {
	valid_buffer := bytes.NewBuffer([]byte{})

	for i := 0; i < b.N; i++ {
		valid_buffer.Write([]byte{
			0x25, 0x04, 0x12, 0x34,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x01,
			0x01,
		})
		ReadPacket(valid_buffer)
	}
}

func TestReadPacket(t *testing.T) {
	Convey("A empty buffer should result in an error", t, func() {
		packet, err := ReadPacket(bytes.NewBuffer([]byte{}))
		So(packet, ShouldBeNil)
		So(err, ShouldEqual, io.EOF)
	})

	Convey("A buffer containing a valid header should result in a packet", t, func() {
		valid_buffer := bytes.NewBuffer([]byte{
			0x25, 0x04, 0x12, 0x34,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x01,
			0x01,
		})

		packet, err := ReadPacket(valid_buffer)
		So(err, ShouldBeNil)
		So(packet, ShouldNotBeNil)
		So(packet.opCode, ShouldEqual, 2)
		So(packet.priority, ShouldEqual, 5)
		So(packet.method, ShouldEqual, 4)
		So(packet.resource, ShouldEqual, int16(0x1234))
		So(packet.sequence, ShouldEqual, 0)
		So(packet.totalSequences, ShouldEqual, 1)
		So(packet.requestId, ShouldResemble, RequestID{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
		So(packet.payloadLength, ShouldEqual, 1)
		So(packet.payload, ShouldResemble, []byte{1})
	})

	Convey("A packet with a payload larger than 1200 bytes should return an error", t, func() {
		valid_buffer := bytes.NewBuffer([]byte{
			0x25, 0x04, 0x12, 0x34,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x01,
			0xFF, 0xFF, 0xFF, 0xFF,
			0x01,
		})
		packet, err := ReadPacket(valid_buffer)
		So(packet, ShouldBeNil)
		So(err, ShouldEqual, PacketPayloadLengthExceeded)
	})
}
