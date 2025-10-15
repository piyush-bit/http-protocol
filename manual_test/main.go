package main

import (
	"fmt"
	"http-protocol/internal/request"
)

func main() {
	data := []byte("GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n")
	r := &request.Request{}
	r.Parse(data)
	fmt.Println(r.Headers["host"])
	fmt.Println(r.Headers["user-agent"])
	fmt.Println(r.Headers["accept"])
}
