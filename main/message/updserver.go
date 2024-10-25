package message

import (
	"echsylon/fudpucker/entity/unit"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/echsylon/go-log"
)

type UdpServer interface {
	Observe(func(string, []byte) error) error
	Send(string, []byte) error
	Stop() error
}

type udpServer struct {
	port       int
	connection net.PacketConn
}

const (
	maxPacketSize   = 10 * unit.MiB
	maxReadTimeout  = 10 * time.Second
	maxWriteTimeout = 5 * time.Second
)

var (
	ErrAlreadyObserving = errors.New("already observing")
	ErrSendFailure      = errors.New("write error")
)

func NewUdpServer(port int) UdpServer {
	return &udpServer{
		port: port,
	}
}

func (s *udpServer) Observe(callback func(string, []byte) error) error {
	if s.connection != nil {
		return ErrAlreadyObserving
	}

	var connectionAddress = fmt.Sprintf(":%d", s.port)
	var connection, err = net.ListenPacket("udp4", connectionAddress)

	if err != nil {
		log.Error("UDP Server failed to open connection [%s]", connectionAddress)
		return err
	} else {
		log.Information("UDP Server opened connection successfully")
		s.connection = connection
		go readBlocking(s.connection, callback)
		return nil
	}
}

func (s *udpServer) Send(address string, data []byte) error {
	if s.connection == nil {
		log.Trace("UDP Server failed send due to connection closed")
		return errors.New("connection closed")
	} else if receiver, err := net.ResolveUDPAddr("udp4", address); err != nil {
		log.Trace("UDP Server failed to resolve receiver address [%s] %s", address, err.Error())
		return err
	} else if _, err := s.connection.WriteTo(data, receiver); err != nil {
		log.Trace("UDP Server failed to write data; [%s] %s", address, err.Error())
		return err
	} else {
		return nil
	}
}

func (s *udpServer) Stop() error {
	if s.connection == nil {
		return nil
	}

	err := s.connection.Close()
	s.connection = nil
	return err
}

func readBlocking(connection net.PacketConn, transport func(string, []byte) error) {
	var buffer = make([]byte, maxPacketSize)
	for {
		// Blocks until either Close() is called on the connection or an
		// error occurs while reading (or waiting for something to read).
		if count, sender, err := connection.ReadFrom(buffer); err != nil {
			if netErr, isNetErr := err.(net.Error); isNetErr && netErr.Timeout() {
				log.Trace("UPD Server read heartbeat")
				continue
			} else if opErr, isOpErr := err.(*net.OpError); isOpErr && opErr.Op == "read" {
				log.Trace("UDP Server connection closed, leaving")
				break
			} else if sender != nil {
				log.Trace("UDP Server failed to read packet from %s, ignoring; %s", sender, err.Error())
				continue
			} else {
				log.Trace("UDP Server failed to read packet, ignoring; %s", err.Error())
				continue
			}
		} else if err := transport(sender.String(), buffer[:count]); err != nil {
			log.Trace("UDP Server failed to process %d bytes from %s, ignoring; %s", count, sender, err.Error())
			continue
		}
	}
}
