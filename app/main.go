package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var filePath = flag.String("directory", "", "directory to serve files")

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

	defer conn.Close()
	buff := make([]byte, 1024)
	_, err := conn.Read(buff)

	if err != nil {
		fmt.Println("Failed to read request")
		os.Exit(1)
	}

	req := string(buff)
	lines := strings.Split(req, CRLF)
	method := strings.Split(lines[0], " ")[0]
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

	case strings.HasPrefix(path, "/files/") && method == "GET":

		fileName, _ := strings.CutPrefix(path, "/files/")

		fileInfo, err := os.Stat(*filePath + "/" + fileName)

		if err != nil {
			res = "HTTP/1.1 404 Not Found" + CRLF + CRLF
			break
		}

		content, _ := os.ReadFile(*filePath + "/" + fileName)
		status := "HTTP/1.1 200 OK" + CRLF
		header := fmt.Sprintf(
			"Content-Type: application/octet-stream"+CRLF+

				"Content-Length: %d"+CRLF+CRLF, fileInfo.Size(),
		)
		res = status + header + string(content)

	case strings.HasPrefix(path, "/files/") && method == "POST":

		fileName, _ := strings.CutPrefix(path, "/files/")

		newPath := filepath.Join(*filePath, fileName)
		newFile, _ := os.Create(newPath)

		defer newFile.Close()

		newFile.WriteString(lines[len(lines)-1])

		status := "HTTP/1.1 201 Created" + CRLF + CRLF
		res = status

	default:
		res = "HTTP/1.1 404 Not Found" + CRLF + CRLF
	}

	conn.Write([]byte(res))
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		fmt.Println("Successful Connection")
		go do(conn)

	}

}
