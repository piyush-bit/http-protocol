package request

import (
	"errors"
	"http-protocol/internal/headers"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       int
}

const (
	INITIAL_STATE = iota
	PARSING_HEADER
	PARSING_BODY
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

func newRequest() *Request {
	return &Request{
		Headers: headers.NewHeaders(),
		State:   INITIAL_STATE,
	}
}

var MALFORMED_REQUEST_LINE_ERROR = errors.New("malformed request line")

func (r *Request) Parse(data []byte) (int, error) {
	readTill := 0
	if r.State == INITIAL_STATE {
		requestLine, i, err := parseRequestLine(string(data))
		if err != nil {
			return 0, err
		}
		if i == 0 {
			return 0, nil
		}
		r.RequestLine = requestLine
		r.State = PARSING_HEADER
		readTill = i
	}

	if r.State == PARSING_HEADER {
		curr := data[readTill:]
		n, done, err := r.Headers.Parse(curr)
		if err != nil {
			return 0, err
		}
		if done {
			r.State = PARSING_BODY
		}
		readTill += n
	}

	if r.State == PARSING_BODY {
		bodyReadTill, done, err := r.ParseBody(data[readTill:])
		if err != nil {
			r.State = DONE_STATE
			return 0, err
		}
		if done {
			r.State = DONE_STATE
		}
		readTill += bodyReadTill
	}
	return readTill, nil
}

func (r *Request) ParseBody(data []byte) (int, bool, error) {
	contentLength, ok := r.Headers.Get("content-length")
	if !ok {
		return 0, true, nil
	}
	contentLengthInt, err := strconv.Atoi(contentLength)
	if err != nil {
		return 0, false, err
	}
	// remaining := contentLengthInt - len(r.Body)
	body := data
	r.Body = append(r.Body, body...)
	if len(r.Body) > contentLengthInt {
		return 0, false, errors.New("body length does not match content-length")
	}
	if len(r.Body) == contentLengthInt {
		return len(body), true, nil
	}
	return len(body), false, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data := []byte{}
	b := make([]byte, 4096)
	n, err := reader.Read(b)
	readTill := 0
	r := newRequest()
	for {
		if err == io.EOF {
			err = nil
			if r.State != DONE_STATE {
				return nil, errors.New("request not completed")
			}
			break
		}
		if err != nil || r.State == DONE_STATE {
			break
		}

		data = append(data, b[:n]...)
		n, err = r.Parse(data[readTill:])
		if err != nil {
			return nil, err
		}
		readTill += n
		n, err = reader.Read(b)
	}

	return r, err
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
	j := strings.Index(requestLineParts[2], "/")
	version := requestLineParts[2][j+1:]

	return RequestLine{
		Method:        requestLineParts[0],
		RequestTarget: requestLineParts[1],
		HttpVersion:   version,
	}, i + 2, nil
}
