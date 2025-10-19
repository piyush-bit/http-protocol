package server

import (
	"bytes"
	"fmt"
	"http-protocol/internal/request"
	"http-protocol/internal/response"
	"io"
	"net"
)

type Server struct {
	Port     int
	Listener net.Listener
}

type HandlerError struct {
	StatusCode int
	Message    string
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func handleConnections(listener net.Listener, handler Handler) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go handleConnection(conn, handler)
	}
}

func handleConnection(conn net.Conn, handler Handler) {
	defer conn.Close()
	req, err := request.RequestFromReader(conn)
	if err != nil {
		return
	}
	var bytebuffer bytes.Buffer
	handlerError := handler(&bytebuffer, req)
	if handlerError != nil {
		HandlerErrorHandler(conn, handlerError)
		return
	}
	err = response.WriteStatusLine(conn, response.OK)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = response.WriteHeaders(conn, response.GetDefaultHeaders(len(bytebuffer.Bytes())))
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = conn.Write(bytebuffer.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}
}

func HandlerErrorHandler(w io.Writer, handlerError *HandlerError) {
	err := response.WriteStatusLine(w, response.StatusCode(handlerError.StatusCode))
	if err != nil {
		fmt.Println(err)
		return
	}
	err = response.WriteHeaders(w, response.GetDefaultHeaders(len(handlerError.Message)))
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = w.Write([]byte(handlerError.Message))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *Server) Close() {
	s.Listener.Close()
}

func Serve(port int, handler Handler) (*Server, error) {
	config := &Server{
		Port:     port,
		Listener: nil,
	}
	listener, err := net.Listen("tcp", ":"+fmt.Sprintf("%d", port))
	if err != nil {
		return nil, err
	}
	config.Listener = listener
	go handleConnections(config.Listener, handler)
	return config, nil
}
