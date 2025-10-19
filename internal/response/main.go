package response

import (
	"fmt"
	"http-protocol/internal/headers"
	"io"
)

type StatusCode int

const (
	OK                    = StatusCode(200)
	NOT_FOUND             = StatusCode(404)
	INTERNAL_SERVER_ERROR = StatusCode(500)
)

var StatusText = map[StatusCode]string{
	OK:                    "200 OK",
	NOT_FOUND:             "404 Not Found",
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
	return headers.Headers{
		"Content-Length": fmt.Sprintf("%d", ContentLen),
		"Content-Type":   "text/plain",
		"Connection":     "close",
	}
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
