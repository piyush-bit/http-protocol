package main

import (
	"fmt"
	"http-protocol/internal/headers"
	"http-protocol/internal/request"
	"http-protocol/internal/response"
	"http-protocol/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"crypto/sha256"
)

const port = 42069

func handler(w *response.Writer, r *request.Request) *server.HandlerError {

	if strings.HasPrefix(r.RequestLine.RequestTarget, "/httpbin") {
		return handleProxy(w, r)
	}
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		w.WriteHeader("Content-Type", "text/html")
		w.WriteStatusLine(response.BAD_REQUEST)
		w.WriteBody([]byte("<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>"))
		return nil
	case "/myproblem":
		w.WriteHeader("Content-Type", "text/html")
		w.WriteStatusLine(response.INTERNAL_SERVER_ERROR)
		w.WriteBody([]byte(`<html>
								<head>
									<title>500 Internal Server Error</title>
								</head>
								<body>
									<h1>Internal Server Error</h1>
									<p>Okay, you know what? This one is on me.</p>
								</body>
								</html>`))
		return nil
	default:
		w.WriteHeader("Content-Type", "text/html")
		w.WriteStatusLine(response.OK)
		w.WriteBody([]byte("<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>"))
		return nil
	}
}

func handleProxy(w *response.Writer, r *request.Request) *server.HandlerError {
	x := r.RequestLine.RequestTarget[9:]
	resp, err := http.Get("http://httpbin.org/" + x)
	if err != nil {
		return &server.HandlerError{
			StatusCode: int(response.INTERNAL_SERVER_ERROR),
			Message:    "Internal Server Error",
		}
	}
	defer resp.Body.Close()
	w.WriteStatusLine(response.OK)
	w.WriteHeaders(response.GetDefaultHeaders(0))
	w.WriteHeader("Transfer-Encoding", "chunked")
	w.WriteHeader("Trailer", "X-Content-SHA256, X-Content-Length")
	w.DeleteHeader("Content-Length")

	buffer := make([]byte, 8)
	data := make([]byte, 0)
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			w.WriteChunkedBody(buffer[:n])
			data = append(data, buffer[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading from proxied request body: %v", err)
			break
		}
	}
	w.WriteChunkedBodyDone()
	w.WriteTrailers(headers.Headers{
		"X-Content-SHA256": fmt.Sprintf("%x", sha256.Sum256(data)),
		"X-Content-Length": fmt.Sprintf("%d", len(data)),
	})
	return nil
}

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
