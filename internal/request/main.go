package request

import (
	"errors"
	"io"
	"regexp"
	"strings"
	"http-protocol/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	State       int
}

const (
	INITIAL_STATE = iota
	PARSING_HEADER
	DONE_STATE
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	SEPARATOR = "\r\n"
)
var MALFORMED_REQUEST_LINE_ERROR = errors.New("malformed request line")


func (r *Request) Parse(data []byte) (int, error) {
	readTill := 0
	switch r.State {
		case INITIAL_STATE:
			requestLine, i, err := parseRequestLine(string(data))
			if err != nil {
				return 0, err
			}
			if i==0 {
				return 0, nil
			}
			r.RequestLine = requestLine
			r.State = PARSING_HEADER
			readTill = i
			return i, nil
		case PARSING_HEADER:
			n, done, err := r.Headers.Parse(data[readTill:])
			if err != nil {
				return 0, err
			}
			if done {
				r.State = DONE_STATE
			}
			readTill += n
			return n, nil
		case DONE_STATE:
			return 0, nil
	}
	return readTill, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data := ""
	b := make([]byte,4096)
	n,err := reader.Read(b)
	r := &Request{}
	for err == nil && r.State != PARSING_HEADER {
		data += string(b[:n])
		n,err = r.Parse([]byte(data))
		if err != nil {
			return nil,err
		}
		n,err = reader.Read(b)
	}

	return r,err
}

// GET /goodies HTTP/1.1

func parseRequestLine(line string) (RequestLine, int, error) {
	i := strings.Index(line, "\r\n")
	if i == -1 {
		return RequestLine{}, 0, nil
	}

	requestLineParts := strings.Split(line[:i], " ")
	if len(requestLineParts) != 3 {
		return RequestLine{}, 0, MALFORMED_REQUEST_LINE_ERROR
	}

	matched, err := regexp.MatchString("^[A-Z]+$", requestLineParts[0])
	if err != nil {
		return RequestLine{}, 0, err
	}
	if !matched {
		return RequestLine{}, 0, MALFORMED_REQUEST_LINE_ERROR
	}

	matched, err = regexp.MatchString("^HTTP/1.1$", requestLineParts[2])
	if err != nil {
		return RequestLine{}, 0, err
	}
	if !matched {
		return RequestLine{}, 0, MALFORMED_REQUEST_LINE_ERROR
	}
	i= strings.Index(requestLineParts[2], "/")
	version := requestLineParts[2][i+1:]

	return RequestLine{
		Method:        requestLineParts[0],
		RequestTarget: requestLineParts[1],
		HttpVersion:   version,
	}, i + 2, nil
}
