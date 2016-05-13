package goryman

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	pb "github.com/golang/protobuf/proto"
	"github.com/bigdatadev/goryman/proto"
)

// Transport is an interface to a generic transport used by the client
type Transport interface {
	SendRecv(message *proto.Msg) (*proto.Msg, error)
	SendMaybeRecv(message *proto.Msg) (*proto.Msg, error)
}

// TcpTransport is a type that implements the Transport interface
type TcpTransport struct {
	conn         net.Conn
	requestQueue chan request
}

// UdpTransport is a type that implements the Transport interface
type UdpTransport struct {
	conn         net.Conn
	requestQueue chan request
}

// request encapsulates a request to send to the Riemann server
type request struct {
	message     *proto.Msg
	response_ch chan response
}

// response encapsulates a response from the Riemann server
type response struct {
	message *proto.Msg
	err     error
}

// MAX_UDP_SIZE is the maximum allowed size of a UDP packet before automatically failing the send
const MAX_UDP_SIZE = 16384

// NewTcpTransport - Factory
func NewTcpTransport(conn net.Conn) *TcpTransport {
	t := &TcpTransport{
		conn:         conn,
		requestQueue: make(chan request),
	}
	go t.runRequestQueue()
	return t
}

// NewUdpTransport - Factory
func NewUdpTransport(conn net.Conn) *UdpTransport {
	t := &UdpTransport{
		conn:         conn,
		requestQueue: make(chan request),
	}
	go t.runRequestQueue()
	return t
}

// TcpTransport implementation of SendRecv, queues a request to send a message to the server
func (t *TcpTransport) SendRecv(message *proto.Msg) (*proto.Msg, error) {
	response_ch := make(chan response)
	t.requestQueue <- request{message, response_ch}
	r := <-response_ch
	return r.message, r.err
}

// TcpTransport implementation of SendMaybeRecv, queues a request to send a message to the server
func (t *TcpTransport) SendMaybeRecv(message *proto.Msg) (*proto.Msg, error) {
	return t.SendRecv(message)
}

// Close will close the TcpTransport
func (t *TcpTransport) Close() error {
	close(t.requestQueue)
	err := t.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// runRequestQueue services the TcpTransport request queue
func (t *TcpTransport) runRequestQueue() {
	for req := range t.requestQueue {
		message := req.message
		response_ch := req.response_ch

		msg, err := t.execRequest(message)

		response_ch <- response{msg, err}
	}
}

// execRequest will send a TCP message to Riemann
func (t *TcpTransport) execRequest(message *proto.Msg) (*proto.Msg, error) {
	msg := &proto.Msg{}
	data, err := pb.Marshal(message)
	if err != nil {
		return msg, err
	}
	b := new(bytes.Buffer)
	if err = binary.Write(b, binary.BigEndian, uint32(len(data))); err != nil {
		return msg, err
	}
	if _, err = t.conn.Write(b.Bytes()); err != nil {
		return msg, err
	}
	if _, err = t.conn.Write(data); err != nil {
		return msg, err
	}
	var header uint32
	if err = binary.Read(t.conn, binary.BigEndian, &header); err != nil {
		return msg, err
	}
	response := make([]byte, header)
	if err = readMessages(t.conn, response); err != nil {
		return msg, err
	}
	if err = pb.Unmarshal(response, msg); err != nil {
		return msg, err
	}
	if msg.GetOk() != true {
		return msg, errors.New(msg.GetError())
	}
	return msg, nil
}

// UdpTransport implementation of SendRecv, will automatically fail if called
func (t *UdpTransport) SendRecv(message *proto.Msg) (*proto.Msg, error) {
	return nil, fmt.Errorf("udp doesn't support receiving acknowledgements")
}

// UdpTransport implementation of SendMaybeRecv, queues a request to send a message to the server
func (t *UdpTransport) SendMaybeRecv(message *proto.Msg) (*proto.Msg, error) {
	response_ch := make(chan response)
	t.requestQueue <- request{message, response_ch}
	r := <-response_ch
	return r.message, r.err
}

// Close will close the UdpTransport
func (t *UdpTransport) Close() error {
	close(t.requestQueue)
	err := t.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// runRequestQueue services the UdpTransport request queue
func (t *UdpTransport) runRequestQueue() {
	for req := range t.requestQueue {
		message := req.message
		response_ch := req.response_ch

		msg, err := t.execRequest(message)

		response_ch <- response{msg, err}
	}
}

// execRequest will send a UDP message to Riemann
func (t *UdpTransport) execRequest(message *proto.Msg) (*proto.Msg, error) {
	data, err := pb.Marshal(message)
	if err != nil {
		return nil, err
	}
	if len(data) > MAX_UDP_SIZE {
		return nil, fmt.Errorf("unable to send message, too large for udp")
	}
	if _, err = t.conn.Write(data); err != nil {
		return nil, err
	}
	return nil, nil
}

// readMessages will read Riemann messages from the TCP connection
func readMessages(r io.Reader, p []byte) error {
	for len(p) > 0 {
		n, err := r.Read(p)
		p = p[n:]
		if err != nil {
			return err
		}
	}
	return nil
}
