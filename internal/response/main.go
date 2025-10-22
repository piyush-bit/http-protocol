package response

import (
	"errors"
	"fmt"
	"http-protocol/internal/headers"
	"io"
)

var DOUBLE_WRITE_ERROR = errors.New("Can't write respoonse twice , use chunked response")
var CHUNKED_BODY_ENDED_ERROR = errors.New("Can't write chunked body after ended")

type Writer struct {
	StatusCode StatusCode
	Headers    headers.Headers
	Writer     io.Writer
	State      int
}

const (	
	INITIAL_STATE = iota
	WRITING_CHUNKED_BODY
	WRITING_TRAILERS
	DONE_STATE
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Writer: w,
		Headers: headers.NewHeaders(),
		State: INITIAL_STATE,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	w.StatusCode = statusCode
	return nil
}

func (w *Writer) DeleteHeader(key string) error {
	delete(w.Headers, key)
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

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.State == DONE_STATE {
		return 0, DOUBLE_WRITE_ERROR
	}
	w.State = DONE_STATE
	err := WriteStatusLine(w.Writer, w.StatusCode)
	if err != nil {
		return 0, err
	}
	w.Headers.Put(GetDefaultHeaders(len(p)))
	err = WriteHeaders(w.Writer, w.Headers)
	if err != nil {
		return 0, err
	}
	return w.Writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.State == DONE_STATE {
		return 0 , CHUNKED_BODY_ENDED_ERROR
	}
	if w.State == INITIAL_STATE{
		w.State = WRITING_CHUNKED_BODY
		err := WriteStatusLine(w.Writer, w.StatusCode)
		if err != nil {
			return 0, err
		}
		w.Headers.Put(GetDefaultHeaders(len(p)))
		w.WriteHeader("Transfer-Encoding", "chunked")
		w.DeleteHeader("Content-Length")
		err = WriteHeaders(w.Writer, w.Headers)
		if err != nil {
			return 0, err
		}
	}
	if w.State == WRITING_CHUNKED_BODY{
		size := len(p)
		fmt.Fprintf(w.Writer,"%x\r\n", size)
		w.Writer.Write(p)
		fmt.Fprint(w.Writer,"\r\n")
	}
	return 0, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error{
	if w.State == WRITING_TRAILERS{
		for header, value := range h {
			fmt.Fprintf(w.Writer,"%s: %s\r\n", header, value)
		}
		fmt.Fprint(w.Writer,"\r\n")
		w.State = DONE_STATE
		return nil
	}
	return errors.New("Trailers not supported")
}
	

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.State == DONE_STATE {
		return 0 , CHUNKED_BODY_ENDED_ERROR
	}
	if w.State == WRITING_CHUNKED_BODY{
		fmt.Fprint(w.Writer,"0\r\n")
		if _, ok := w.Headers.Get("Trailer"); !ok {
			fmt.Fprint(w.Writer,"\r\n")
			w.State = DONE_STATE
		}else{
			w.State = WRITING_TRAILERS
		}
	}
	return 0, nil
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
