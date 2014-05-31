package teaspoon

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
)

func TestReadPacket(t *testing.T) {
	Convey("A empty packet should result in an error", t, func() {
		packet, err := ReadPacket(bytes.NewBuffer([]byte{}))
		So(packet, ShouldBeNil)
		So(err, ShouldEqual, io.EOF)
	})

	Convey("A valid header should result in a packet", t, func() {
		valid_packet := bytes.NewBuffer([]byte{
			0x25, 0x04, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x01,
			0x01,
		})

		packet, err := ReadPacket(valid_packet)
		So(packet, ShouldNotBeNil)
		So(err, ShouldBeNil)
		So(packet.opCode, ShouldEqual, 2)
		So(packet.priority, ShouldEqual, 5)
		So(packet.sequence, ShouldEqual, 0)
		So(packet.totalSequences, ShouldEqual, 1)
		So(packet.requestId, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
		So(packet.payloadLength, ShouldEqual, 1)
		So(packet.payload, ShouldResemble, []byte{1})
	})
}
