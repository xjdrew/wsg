package main

import (
	"net"
)

type ServerConn struct {
	net.Conn
}

type Server struct {
	addr string
}

func (server Server) Dial() (*ServerConn, error) {
	conn, err := net.Dial("tcp", server.addr)
	if err != nil {
		return nil, err
	}
	return &ServerConn{conn}, nil
}

func newServer(addr string) *Server {
	return &Server{addr: addr}
}
