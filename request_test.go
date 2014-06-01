package teaspoon

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"testing"
)

func TestConstructRequest(t *testing.T) {
	Convey("A nil packet slice should result in an error", t, func() {
		request, err := constructRequest(nil, nil)
		So(request, ShouldBeNil)
		So(err, ShouldEqual, InvalidPacketSequence)
	})

	Convey("A nil request ID should result in an error", t, func() {
		request, err := constructRequest([]*Packet{}, nil)
		So(request, ShouldBeNil)
		So(err, ShouldEqual, InvalidRequestId)
	})

	Convey("An array of packets that don't satisfy the request ID should return an error", t, func() {
		request, err := constructRequest([]*Packet{}, []byte{1})
		So(request, ShouldBeNil)
		So(err, ShouldEqual, RequestNotReady)
	})

	Convey("An array of packets that DO satisfy the request ID should return a request", t, func() {
		requestId := []byte{1}
		packets := []*Packet{
			&Packet{
				opCode:         2,
				priority:       5,
				sequence:       0,
				totalSequences: 1,
				requestId:      requestId,
				payloadLength:  1,
				payload:        []byte{1},
			},
		}

		request, err := constructRequest(packets, []byte{1})
		So(request, ShouldNotBeNil)
		So(err, ShouldBeNil)

		So(request.RequestID, ShouldResemble, requestId)
		So(request.Payload, ShouldResemble, []byte{1})
	})
}

func TestReadRequest(t *testing.T) {
	Convey("A empty buffer should result in an error", t, func() {
		request, err := ReadRequest(bytes.NewBuffer([]byte{}))
		So(request, ShouldBeNil)
		So(err, ShouldEqual, io.EOF)
	})

	Convey("A valid buffer with a single packet should result in an request", t, func() {
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

		request, err := ReadRequest(valid_buffer)
		So(request, ShouldNotBeNil)
		So(err, ShouldBeNil)
		So(request.Method, ShouldEqual, 4)
		So(request.Resource, ShouldEqual, int16(0x1234))
		So(request.RequestID, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1})
		So(request.Payload, ShouldResemble, []byte{1})
	})

	Convey("A valid buffer with multiple packets should result in an request", t, func() {
		valid_buffer := bytes.NewBuffer([]byte{
			// First Packet In Sequence
			0x25, 0x04, 0x12, 0x34,
			0x00, 0x00, 0x00, 0x02,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x01, 0x02,
			0x00, 0x00, 0x00, 0x01,
			0x01,
			// Second Packet In Sequence
			0x25, 0x04, 0x12, 0x34,
			0x00, 0x01, 0x00, 0x02,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x01, 0x02,
			0x00, 0x00, 0x00, 0x02,
			0x02, 0x03,
		})

		request, err := ReadRequest(valid_buffer)
		So(request, ShouldNotBeNil)
		So(err, ShouldBeNil)
		So(request.Method, ShouldEqual, 4)
		So(request.Resource, ShouldEqual, int16(0x1234))
		So(request.RequestID, ShouldResemble, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2})
		So(request.Payload, ShouldResemble, []byte{1, 2, 3})
	})
}
