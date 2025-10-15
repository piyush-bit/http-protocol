package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main(){
	f,err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := net.DialUDP("udp", nil, f)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(">")
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		conn.Write([]byte(message))
		fmt.Printf(">")
	}
	
}