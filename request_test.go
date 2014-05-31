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
}
