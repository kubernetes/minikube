// A Riemann client for Go, featuring concurrency, sending events and state updates, queries,
// and feature parity with the reference implementation written in Ruby.
//
// Copyright (C) 2014 by Christopher Gilbert <christopher.john.gilbert@gmail.com>
package goryman

import (
	"net"
	"time"

	pb "github.com/golang/protobuf/proto"
	"github.com/bigdatadev/goryman/proto"
)

// GorymanClient is a client library to send events to Riemann
type GorymanClient struct {
	udp  *UdpTransport
	tcp  *TcpTransport
	addr string
}

// NewGorymanClient - Factory
func NewGorymanClient(addr string) *GorymanClient {
	return &GorymanClient{
		addr: addr,
	}
}

// Connect creates a UDP and TCP connection to a Riemann server
func (c *GorymanClient) Connect() error {
	udp, err := net.DialTimeout("udp", c.addr, time.Second*5)
	if err != nil {
		return err
	}
	tcp, err := net.DialTimeout("tcp", c.addr, time.Second*5)
	if err != nil {
		return err
	}
	c.udp = NewUdpTransport(udp)
	c.tcp = NewTcpTransport(tcp)
	return nil
}

// Close the connection to Riemann
func (c *GorymanClient) Close() error {
	if nil == c.udp && nil == c.tcp {
		return nil
	}
	err := c.udp.Close()
	if err != nil {
		return err
	}
	return c.tcp.Close()
}

// Send an event
func (c *GorymanClient) SendEvent(e *Event) error {
	epb, err := EventToProtocolBuffer(e)
	if err != nil {
		return err
	}

	message := &proto.Msg{}
	message.Events = append(message.Events, epb)

	_, err = c.sendMaybeRecv(message)
	return err
}

// Send a state update
func (c *GorymanClient) SendState(s *State) error {
	spb, err := StateToProtocolBuffer(s)
	if err != nil {
		return err
	}

	message := &proto.Msg{}
	message.States = append(message.States, spb)

	_, err = c.sendMaybeRecv(message)
	return err
}

// Query the server for events
func (c *GorymanClient) QueryEvents(q string) ([]Event, error) {
	query := &proto.Query{}
	query.String_ = pb.String(q)

	message := &proto.Msg{}
	message.Query = query

	response, err := c.sendRecv(message)
	if err != nil {
		return nil, err
	}

	return ProtocolBuffersToEvents(response.GetEvents()), nil
}

// Send and receive data from Riemann
func (c *GorymanClient) sendRecv(m *proto.Msg) (*proto.Msg, error) {
	return c.tcp.SendRecv(m)
}

// Send and maybe receive data from Riemann
func (c *GorymanClient) sendMaybeRecv(m *proto.Msg) (*proto.Msg, error) {
	_, err := c.udp.SendMaybeRecv(m)
	if err != nil {
		return c.tcp.SendMaybeRecv(m)
	}
	return nil, nil
}
