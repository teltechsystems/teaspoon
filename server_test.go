package teaspoon

import (
	"bytes"
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"io"
	"net"
	"testing"
	"time"
)

// type dummyAddr struct {
// 	network string
// }

// func (a *dummyAddr) Network() string {
// 	return a.network
// }

// func (a *dummyAddr) String() string {
// 	return a.network
// }

type dummyConn struct {
	io.Reader
	io.Writer
	closed bool
}

func (c *dummyConn) Close() error {
	c.closed = true

	return nil
}

func (c *dummyConn) LocalAddr() net.Addr {
	return nil
}

func (c *dummyConn) RemoteAddr() net.Addr {
	return nil
}

func (c *dummyConn) SetDeadline(t time.Time) error {
	return nil
}

func (c *dummyConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (c *dummyConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type dummyListener struct {
	conns  []*dummyConn
	index  int
	closed bool
}

func (l *dummyListener) Accept() (net.Conn, error) {
	if l.index >= len(l.conns) {
		return nil, &net.OpError{"read", "tcp", l.Addr(), errors.New("Connection closed")}
	}

	conn := l.conns[l.index]

	l.index += 1

	return conn, nil
}

func (l *dummyListener) Addr() net.Addr {
	return nil
}

func (l *dummyListener) Close() error {
	l.closed = true
	return nil
}

func TestConnServe(t *testing.T) {
	Convey("With a valid connection and an empty buffer, conn should close immediately", t, func() {
		server := &Server{Handler: nil}
		reader := bytes.NewBuffer([]byte{})
		writer := bytes.NewBuffer([]byte{})

		rwc := &dummyConn{Reader: reader, Writer: writer}
		conn := &conn{rwc: rwc, srv: server}
		conn.serve()

		So(writer.Bytes(), ShouldResemble, []byte{})
		So(rwc.closed, ShouldBeTrue)
	})

	Convey("With a valid connection a valid buffer, handler should be called", t, func() {
		handler_called := make(chan bool, 1)
		handler := HandlerFunc(func(w ResponseWriter, r *Request) {
			handler_called <- true
			w.Write([]byte("HELLO"))
		})
		server := &Server{Handler: handler}
		reader := bytes.NewBuffer([]byte{
			0x85, 0x04, 0x12, 0x34,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x01,
			0x12,
		})
		writer := bytes.NewBuffer([]byte{})

		rwc := &dummyConn{Reader: reader, Writer: writer}
		conn := &conn{rwc: rwc, srv: server}
		conn.serve()

		So(<-handler_called, ShouldBeTrue)
		So(writer.Bytes(), ShouldResemble, []byte{
			0x85, 0x01, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x05,
			72, 69, 76, 76, 79, // Hello String
		})
		So(rwc.closed, ShouldBeTrue)
	})
}

func TestServerServe(t *testing.T) {
	server := &Server{}

	reader := bytes.NewBuffer([]byte("HELLO"))
	writer := bytes.NewBuffer([]byte{})

	Convey("With a valid listener that returns a single connection, all connections should be accepted", t, func() {
		conns := []*dummyConn{&dummyConn{Reader: reader, Writer: writer}}

		listener := &dummyListener{conns: conns}

		So(listener.index, ShouldEqual, 0)
		server.Serve(listener)
		So(listener.index, ShouldEqual, 1)

		<-time.After(time.Millisecond * 1)

		So(listener.closed, ShouldBeTrue)
		So(listener.conns[0].closed, ShouldBeTrue)
	})

	Convey("With a valid listener that returns a multiple connections, all connections should be accepted", t, func() {
		conns := []*dummyConn{
			&dummyConn{Reader: reader, Writer: writer},
			&dummyConn{Reader: reader, Writer: writer},
		}

		listener := &dummyListener{conns: conns}

		So(listener.index, ShouldEqual, 0)
		server.Serve(listener)
		So(listener.index, ShouldEqual, 2)

		<-time.After(time.Millisecond * 1)

		So(listener.closed, ShouldBeTrue)
		So(listener.conns[0].closed, ShouldBeTrue)
		So(listener.conns[1].closed, ShouldBeTrue)
	})
}
