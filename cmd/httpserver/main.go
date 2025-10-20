package main

import (
	"http-protocol/internal/request"
	"http-protocol/internal/response"
	"http-protocol/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func handler(w *response.Writer, r *request.Request) *server.HandlerError {
	switch r.RequestLine.RequestTarget {
	case "/yourproblem":
		w.WriteBody([]byte("<html><head><title>400 Bad Request</title></head><body><h1>Bad Request</h1><p>Your request honestly kinda sucked.</p></body></html>"))
		w.WriteHeader("Content-Type", "text/html")
		w.WriteStatusLine(response.BAD_REQUEST)
		return nil
	case "/myproblem":
		w.WriteBody([]byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`))
		w.WriteHeader("Content-Type", "text/html")
		w.WriteStatusLine(response.INTERNAL_SERVER_ERROR)
		return nil
	default:
		w.WriteBody([]byte("<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>"))
		w.WriteHeader("Content-Type", "text/html")
		w.WriteStatusLine(response.OK)
		return nil
	}
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
