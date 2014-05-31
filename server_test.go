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
	reader := bytes.NewBuffer([]byte{})
	writer := bytes.NewBuffer([]byte{})

	Convey("With a valid connection, serve should be usable", t, func() {
		conn := &conn{rwc: &dummyConn{Reader: reader, Writer: writer}}
		conn.serve()
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

		So(listener.closed, ShouldBeTrue)
		So(listener.conns[0].closed, ShouldBeTrue)
		So(listener.conns[1].closed, ShouldBeTrue)
	})
}
