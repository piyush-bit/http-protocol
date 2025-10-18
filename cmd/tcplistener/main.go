package main

import (
	"fmt"
	"net"
	"http-protocol/internal/request"
)

func main() {
	f, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	for {
		conn, err := f.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		r,err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Request line:\n- Method : %s\n- Target : %s\n- Version : %s\nHeaders : %v\nBody : %s\n", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion, r.Headers, r.Body)
	}
}

