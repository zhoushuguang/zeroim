package socketio

import (
	"net"

	"github.com/zhoushuguang/zeroim/common/libnet"
)

type Server struct {
	Name         string
	Address      string
	Manager      *libnet.Manager
	Protocol     libnet.Protocol
	SendChanSize int
}

func NewServe(name, address string, protocol libnet.Protocol, sendChanSize int) (*Server, error) {
	return &Server{
		Name:         name,
		Address:      address,
		Manager:      libnet.NewManager(name),
		Protocol:     protocol,
		SendChanSize: sendChanSize,
	}, nil
}

func (s *Server) Accept(conn net.Conn) (*libnet.Session, error) {
	return libnet.NewSession(
		s.Manager,
		s.Protocol.NewCodec(conn),
		s.SendChanSize,
	), nil
}
