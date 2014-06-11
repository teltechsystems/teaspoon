package teaspoon

import (
	"bytes"
	"io"
	"log"
	"net"
	"os"
)

const (
	OPCODE_CONTINUATION = 0x0
	OPCODE_TEXT         = 0x1
	OPCODE_BINARY       = 0x2
	OPCODE_CLOSE        = 0x8
	OPCODE_PING         = 0x9
	OPCODE_PONG         = 0xA
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
	r.reply.Resource = resource
}

func (r *response) Write(p []byte) (int, error) {
	return r.w.Write(p)
}

func (r *response) finishRequest() {
	MAX_MTU := int32(1200)
	payload := r.w.Bytes()
	payloadLength := len(payload)
	totalSequences := int32(payloadLength)/MAX_MTU + 1

	logger.Printf("finishRequest - payload : %v", payload)
	logger.Printf("finishRequest - payloadLength : %v", payloadLength)
	logger.Printf("finishRequest - totalSequences : %v", totalSequences)

	for sequence := int32(0); sequence < totalSequences; sequence++ {
		r.conn.rwc.Write([]byte{
			(r.reply.OpCode << 4) | r.req.Priority, r.reply.Method, byte(r.reply.Resource >> 8), byte(r.reply.Resource),
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
		logger.Printf("finishRequest - chunk written : %v", payload[sequence*MAX_MTU:sequence*MAX_MTU+payloadLength])
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
		conn:  c,
		req:   req,
		reply: &Request{OpCode: OPCODE_BINARY, Method: 0x01, Resource: 0x00},
		w:     bytes.NewBuffer([]byte{}),
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

		switch int(responseWriter.req.OpCode) {
		case OPCODE_PING:
			responseWriter.reply.OpCode = OPCODE_PONG
			responseWriter.finishRequest()
		default:
			c.srv.Handler.ServeTSP(responseWriter, responseWriter.req)
			responseWriter.finishRequest()
		}
	}
}
