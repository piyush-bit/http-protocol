package response

import (
	"fmt"
	"http-protocol/internal/headers"
	"io"
)

type Writer struct {
	StatusCode StatusCode
	Headers    headers.Headers
	Body       []byte
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	w.StatusCode = statusCode
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	w.Headers.Put(headers)
	return nil
}

func (w *Writer) WriteHeader(key string, value string) error {
	w.Headers.Put(headers.Headers{
		key: value,
	})
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error){
	w.Body = append(w.Body, p...)
	w.WriteHeader("Content-Length", fmt.Sprintf("%d", len(w.Body)))
	return len(p), nil
}

type StatusCode int

const (
	OK                    = StatusCode(200)
	NOT_FOUND             = StatusCode(404)
	BAD_REQUEST           = StatusCode(400)
	INTERNAL_SERVER_ERROR = StatusCode(500)
)

var StatusText = map[StatusCode]string{
	OK:                    "200 OK",
	NOT_FOUND:             "404 Not Found",
	BAD_REQUEST:           "400 Bad Request",
	INTERNAL_SERVER_ERROR: "500 Internal Server Error",
}

func WriteStatusLine(w io.Writer, status StatusCode) error {
	_, err := fmt.Fprintf(w, "HTTP/1.1 %d %s\r\n", status, StatusText[status])
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultHeaders(ContentLen int) headers.Headers {
	header := headers.NewHeaders()
	header.Put(headers.Headers{
		"Content-Length": fmt.Sprintf("%d", ContentLen),
		"Content-Type":   "text/plain",
		"Connection":     "close",
	})
	return header
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for header, value := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", header, value)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "\r\n")
	if err != nil {
		return err
	}
	return nil
}
