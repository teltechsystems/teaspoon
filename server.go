package teaspoon

import (
	"bytes"
	"io"
	"log"
	"net"
	"os"
	"runtime/debug"
)

const (
	OPCODE_CONTINUATION = 0x0
	OPCODE_TEXT         = 0x1
	OPCODE_BINARY       = 0x2
	OPCODE_CLOSE        = 0x8
	OPCODE_PING         = 0x9
	OPCODE_PONG         = 0xA
	CLIENT_CONNECT      = 1
	CLIENT_DISCONNECT   = 2
)

var (
	logger = log.New(os.Stdout, "[teaspoon] ", 0)
)

type Handler interface {
	ServeTSP(ResponseWriter, *Request)
}

type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeTSP(w ResponseWriter, r *Request) {
	f(w, r)
}

type ResponseWriter interface {
	SetResource(int)
	GetDirectWriter() io.Writer
	Write([]byte) (int, error)
}

type Server struct {
	Addr    string
	Handler Handler
	binders []Binder
}

func (srv *Server) AddBinder(b Binder) {
	srv.binders = append(srv.binders, b)
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
	l, e := net.Listen("tcp", addr)
	if e != nil {
		return e
	}
	return srv.Serve(l)
}

func (s *Server) Serve(l net.Listener) error {
	defer l.Close()

	for {
		rwc, err := l.Accept()
		if err != nil {
			return err
		}

		conn := &conn{
			rwc:       rwc,
			srv:       s,
			frameChan: make(chan []byte, 10),
			quitChan:  make(chan bool),
		}
		go conn.serve()
	}

	return nil
}

func (s *Server) triggerEvent(eventType int, c io.Writer) {
	for i := range s.binders {
		switch eventType {
		case CLIENT_CONNECT:
			s.binders[i].OnClientConnect(c)
		case CLIENT_DISCONNECT:
			s.binders[i].OnClientDisconnect(c)
		}
	}
}

func ListenAndServe(addr string, handler Handler) error {
	server := &Server{Addr: addr, Handler: handler}
	return server.ListenAndServe()
}

type response struct {
	conn  *conn
	req   *Request
	reply *Request
	w     *bytes.Buffer
}

func (r *response) GetDirectWriter() io.Writer {
	return r.conn
}

func (r *response) SetResource(resource int) {
	r.reply.Resource = resource
}

func (r *response) Write(p []byte) (int, error) {
	return r.w.Write(p)
}

func (r *response) finishRequest() {
	r.reply.RequestID = r.req.RequestID
	r.reply.Priority = r.req.Priority
	r.reply.Payload = r.w.Bytes()
	r.reply.WriteTo(r.conn)
}

type conn struct {
	rwc       io.ReadWriteCloser
	srv       *Server
	frameChan chan []byte
	quitChan  chan bool
}

func (c *conn) readRequest(r io.Reader) (*response, error) {
	req, err := ReadRequest(r)
	if err != nil {
		return nil, err
	}

	return &response{
		conn:  c,
		req:   req,
		reply: &Request{OpCode: OPCODE_BINARY, Method: 0x01, Resource: 0x00},
		w:     bytes.NewBuffer([]byte{}),
	}, nil
}

func (c *conn) Write(p []byte) (int, error) {
	c.frameChan <- p

	return len(p), nil
}

func (c *conn) serve() {
	c.srv.triggerEvent(CLIENT_CONNECT, c)
	defer c.srv.triggerEvent(CLIENT_DISCONNECT, c)

	defer c.rwc.Close()
	defer func() {
		if r := recover(); r != nil {
			logger.Printf("Recovered client crash: %s", r)
			debug.PrintStack()
		}
	}()

	go func() {
		for {
			responseWriter, err := c.readRequest(c.rwc)
			if err != nil {
				logger.Printf("Error reading:", err)
				c.quitChan <- true
				return
			}

			switch int(responseWriter.req.OpCode) {
			case OPCODE_PING:
				responseWriter.reply.OpCode = OPCODE_PONG
				responseWriter.finishRequest()
			default:
				c.srv.Handler.ServeTSP(responseWriter, responseWriter.req)
				responseWriter.finishRequest()
			}
		}
	}()

	for {
		select {
		case <-c.quitChan:
			return
		case frame := <-c.frameChan:
			logger.Println("Writing a frame:", frame)
			bytesWritten, err := c.rwc.Write(frame)
			logger.Println("Wanted to write", len(frame), "bytes. We ended up writing", bytesWritten, "with this error", err)
		}
	}
}
