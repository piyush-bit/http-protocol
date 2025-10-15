package headers

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

const SEPARATOR = "\r\n"

var INVALID_HEADER_FORMAT_ERROR = errors.New("invalid header format")
var INVALID_KEY_FORMAT_ERROR = errors.New("invalid key format")

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	i := 0
	for {
		j := bytes.Index(data[i:], []byte(SEPARATOR))
		if j == -1 {
			break // not enough data yet
		}

		// Detect end of headers (\r\n\r\n)
		if j == 0 {
			return i + 2, true, nil
		}

		line := data[i : i+j]
		columnIndex := bytes.Index(line, []byte(":"))
		if columnIndex == -1 {
			return i, false, INVALID_HEADER_FORMAT_ERROR
		}
		headerKey := string(line[:columnIndex])
		if strings.HasSuffix(headerKey, " ") {
			return i, false, INVALID_HEADER_FORMAT_ERROR
		}
		headerKey = strings.TrimSpace(headerKey)
		matched, err := regexp.MatchString("^[a-zA-Z0-9 !,#$%&'*+-.^_`|~]+$", headerKey)
		if err != nil {
			return i, false, err
		}
		if !matched {
			return i, false, INVALID_KEY_FORMAT_ERROR
		}
		headerValue := strings.TrimSpace(string(line[columnIndex+1:]))
		if headerValue == "" {
			return i, false, INVALID_HEADER_FORMAT_ERROR
		}

		// additional header parsing
		fmt.Println("Addig Header: ", headerKey, headerValue)

		if _, ok := h[strings.ToLower(headerKey)]; ok {
			h[strings.ToLower(headerKey)] = h[strings.ToLower(headerKey)] + ", " + headerValue
		} else {
			h[strings.ToLower(headerKey)] = headerValue
		}
		i += j + 2
	}

	

	return i, false, nil
}
