package binders

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/teltechsystems/teaspoon"
	"io"
	"math/rand"
	"testing"
	"time"
)

type dummyRwc struct {
	io.ReadWriter
}

func (d *dummyRwc) Close() error {
	return nil
}

func newDummyRwc(rw io.ReadWriter) io.ReadWriteCloser {
	return &dummyRwc{rw}
}

func TestNewPinger(t *testing.T) {
	Convey("A basic pinger should set up a few known properties", t, func() {
		pinger := NewPinger(time.Second * 1)
		So(pinger.interval, ShouldEqual, time.Second*1)
		So(len(pinger.rwcs), ShouldEqual, 0)

		_, matchesInterface := interface{}(pinger).(teaspoon.Binder)
		So(matchesInterface, ShouldEqual, true)

		rwc := newDummyRwc(bytes.NewBuffer([]byte{}))

		pinger.OnClientConnect(rwc)
		So(len(pinger.rwcs), ShouldEqual, 1)
		So(pinger.rwcs[0], ShouldEqual, rwc)

		pinger.OnClientDisconnect(rwc)
		So(len(pinger.rwcs), ShouldEqual, 1)
		So(pinger.rwcs[0], ShouldBeNil)
	})
}

func TestPingerSendPing(t *testing.T) {
	Convey("With a valid pinger delivering static pings", t, func() {
		rand.Seed(0)
		pinger := NewPinger(time.Second * 1000)

		buffer := bytes.NewBuffer([]byte{})
		rwc := newDummyRwc(buffer)
		pinger.OnClientConnect(rwc)
		pinger.sendPing(rwc)

		So(buffer.Bytes(), ShouldResemble, []byte{149, 0, 0, 0, 0, 0, 0, 1, 10, 2, 9, 10, 11, 0, 15, 5, 8, 0, 8, 11, 11, 12, 8, 14, 0, 0, 0, 0})
	})
}

func TestPingerProcessPings(t *testing.T) {
	Convey("With the pingProcessor running at known intervals", t, func() {
		rand.Seed(0)
		pinger := NewPinger(time.Second * 1)

		buffer := bytes.NewBuffer([]byte{})
		rwc := newDummyRwc(buffer)
		pinger.OnClientConnect(rwc)
		time.Sleep(time.Second * 1)

		So(buffer.Bytes(), ShouldResemble, []byte{149, 0, 0, 0, 0, 0, 0, 1, 10, 2, 9, 10, 11, 0, 15, 5, 8, 0, 8, 11, 11, 12, 8, 14, 0, 0, 0, 0})
		buffer.Reset()

		time.Sleep(time.Second * 1)
		So(buffer.Bytes(), ShouldResemble, []byte{149, 0, 0, 0, 0, 0, 0, 1, 15, 2, 2, 14, 11, 4, 10, 12, 1, 10, 0, 0, 15, 12, 13, 4, 0, 0, 0, 0})
	})
}
