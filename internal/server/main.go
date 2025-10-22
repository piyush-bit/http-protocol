package server

import (
	"fmt"
	"http-protocol/internal/request"
	"http-protocol/internal/response"
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

type Handler func(r *response.Writer, req *request.Request) *HandlerError

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
	writer := response.NewWriter(conn)
	handlerError := handler(writer, req)
	if handlerError != nil {
		HandlerErrorHandler(writer, handlerError)
		return
	}
}

func HandlerErrorHandler(w *response.Writer, handlerError *HandlerError) {
	w.StatusCode = response.StatusCode(handlerError.StatusCode)
	w.WriteHeaders(response.GetDefaultHeaders(len(handlerError.Message)))
	_, err := w.WriteBody([]byte(handlerError.Message))
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
