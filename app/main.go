package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
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
	buffer := make([]byte, 4096)
	data := ""

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Read error:", err)
			}
			return
		}

		data += string(buffer[:n])

		// Handle multiple requests in case they come together
		for {
			// Find end of headers
			headerEnd := strings.Index(data, CRLF+CRLF)
			if headerEnd == -1 {
				break
			}

			request := data[:headerEnd+4]
			lines := strings.Split(request, CRLF)
			requestLine := strings.Split(lines[0], " ")
			method := requestLine[0]
			path := requestLine[1]
			headerMap := headerToMap(lines)

			// Determine if body is expected
			contentLength := 0
			if cl, ok := headerMap["Content-Length"]; ok {
				contentLength, _ = strconv.Atoi(cl)
			}

			// Check if full request (headers + body) has arrived
			totalLength := headerEnd + 4 + contentLength
			if len(data) < totalLength {
				break
			}

			body := data[headerEnd+4 : totalLength]
			data = data[totalLength:] // Remove processed request

			// Build response
			var res string
			switch {
			case path == "/":
				res = "HTTP/1.1 200 OK" + CRLF + CRLF

			case strings.HasPrefix(path, "/echo/"):
				bodyText, _ := strings.CutPrefix(path, "/echo/")
				status := "HTTP/1.1 200 OK" + CRLF
				header := fmt.Sprintf(
					"Content-Type: text/plain"+CRLF+
						"Content-Length: %d"+CRLF+CRLF, len(bodyText),
				)
				res = status + header + bodyText

			case strings.HasPrefix(path, "/user-agent"):
				ua := headerMap["User-Agent"]
				status := "HTTP/1.1 200 OK" + CRLF
				header := fmt.Sprintf(
					"Content-Type: text/plain"+CRLF+
						"Content-Length: %d"+CRLF+CRLF, len(ua),
				)
				res = status + header + ua

			case strings.HasPrefix(path, "/files/") && method == "GET":
				fileName, _ := strings.CutPrefix(path, "/files/")
				filePath := filepath.Join(*filePath, fileName)
				fileInfo, err := os.Stat(filePath)
				if err != nil {
					res = "HTTP/1.1 404 Not Found" + CRLF + CRLF
					break
				}
				content, _ := os.ReadFile(filePath)
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
				newFile.WriteString(body)
				res = "HTTP/1.1 201 Created" + CRLF + CRLF

			default:
				res = "HTTP/1.1 404 Not Found" + CRLF + CRLF
			}

			// Check if client wants to close connection
			if connHeader, ok := headerMap["Connection"]; ok && strings.ToLower(connHeader) == "close" {
				conn.Write([]byte(res))
				return
			}

			conn.Write([]byte(res))
		}
	}
}

func main() {
	flag.Parse()
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}
		go do(conn)
	}
}
