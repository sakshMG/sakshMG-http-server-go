package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const CRLF = "\r\n"

func headerToMap(lines []string) map[string]string {

	mp := make(map[string]string)

	for _, line := range lines[1:] {
		if line == "" {
			break
		}

		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			key := strings.TrimSpace(headerParts[0])
			value := strings.TrimSpace(headerParts[1])
			mp[key] = value
		}
	}
	return mp

}

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

	headerMap := headerToMap(lines)

	var res string

	switch {

	case path == "/":
		res = "HTTP/1.1 200 OK" + CRLF + CRLF

	case strings.HasPrefix(path, "/echo/"):

		body, _ := strings.CutPrefix(path, "/echo/")

		status := "HTTP/1.1 200 OK" + CRLF
		header := fmt.Sprintf(
			"Content-Type: text/plain"+CRLF+

				"Content-Length: %d"+CRLF+CRLF, len(body),
		)
		res = status + header + body

	case strings.HasPrefix(path, "/user-agent"):

		body := headerMap["User-Agent"]

		status := "HTTP/1.1 200 OK" + CRLF
		header := fmt.Sprintf(
			"Content-Type: text/plain"+CRLF+

				"Content-Length: %d"+CRLF+CRLF, len(body),
		)

		res = status + header + body

	default:
		res = "HTTP/1.1 404 Not Found" + CRLF + CRLF
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
