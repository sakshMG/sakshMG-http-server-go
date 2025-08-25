package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const CRLF = "\r\n"

func do(conn net.Conn) {

	buff := make([]byte, 1024)
	_, err := conn.Read(buff)

	if err != nil {
		fmt.Println("Failed to read request")
		os.Exit(1)
	}

	req := string(buff)
	lines := strings.Split(req, CRLF)
	path := strings.Split(lines[0], " ")[1]

	var res string

	if path == "/" {
		res = "HTTP/1.1 200 OK\r\n\r\n"
	} else {
		res = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	conn.Write([]byte(res))
	conn.Close()
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	do(conn)
}
