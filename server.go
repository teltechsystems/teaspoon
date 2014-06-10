package teaspoon

import (
	"bytes"
	"io"
	"log"
	"net"
	"os"
)

var logger = log.New(os.Stdout, "[teaspoon] ", 0)

type Handler interface {
	ServeTSP(ResponseWriter, *Request)
}

type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeTSP(w ResponseWriter, r *Request) {
	f(w, r)
}

type ResponseWriter interface {
	SetResource(int)
	Write([]byte) (int, error)
}

type Server struct {
	Addr    string
	Handler Handler
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

		conn := &conn{rwc: rwc, srv: s}
		go conn.serve()
	}

	return nil
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

func (r *response) SetResource(resource int) {
	if r.reply == nil {
		r.reply = &Request{Method: 0x01, Resource: 0x00}
	}

	r.reply.Resource = resource
}

func (r *response) Write(p []byte) (int, error) {
	return r.w.Write(p)
}

func (r *response) finishRequest() {
	MAX_MTU := int32(1200)
	payload := r.w.Bytes()
	totalSequences := int32(len(payload))/MAX_MTU + 1

	if r.reply == nil {
		r.reply = &Request{Method: 0x01, Resource: 0x00}
	}

	for sequence := int32(0); sequence < totalSequences; sequence++ {
		r.conn.rwc.Write([]byte{
			0x80 | r.req.Priority, r.reply.Method, byte(r.reply.Resource >> 8), byte(r.reply.Resource),
			byte(sequence >> 8), byte(sequence), byte(totalSequences >> 8), byte(totalSequences),
		})

		// We must ensure the request ID matches the original request
		r.conn.rwc.Write(r.req.RequestID[:])

		payloadLength := MAX_MTU
		if int32(len(payload)) < (sequence+1)*MAX_MTU {
			payloadLength = int32(len(payload)) - sequence*MAX_MTU
		}

		r.conn.rwc.Write([]byte{
			byte(payloadLength >> 24), byte(payloadLength >> 16), byte(payloadLength >> 8), byte(payloadLength),
		})

		r.conn.rwc.Write(payload[sequence*MAX_MTU : sequence*MAX_MTU+payloadLength])
	}
}

type conn struct {
	rwc io.ReadWriteCloser
	srv *Server
}

func (c *conn) readRequest(r io.Reader) (*response, error) {
	req, err := ReadRequest(r)
	if err != nil {
		return nil, err
	}

	return &response{
		conn: c,
		req:  req,
		w:    bytes.NewBuffer([]byte{}),
	}, nil
}

func (c *conn) serve() {
	defer c.rwc.Close()

	for {
		responseWriter, err := c.readRequest(c.rwc)
		if err != nil {
			logger.Printf("Error reading:", err)
			return
			// if err == io.EOF {
			// 	return
			// }
		}

		c.srv.Handler.ServeTSP(responseWriter, responseWriter.req)
		responseWriter.finishRequest()
	}
}
