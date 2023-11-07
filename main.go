package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	// "math/rand"
	// "sync"
)

type Server struct {
	port int
	pool chan net.Conn
	// mu sync.Mutex
}

func extractRequestFields(conn net.Conn) []string {
	buffer := make([]byte, 1024)
	mLen, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	return strings.Fields(string(buffer[:mLen]))
}

func handleGetRequest(conn net.Conn, requestFields []string) http.Response {
	pwd, _ := os.Getwd()
	filename := requestFields[1]
	file, err := os.ReadFile(pwd + filename)
	var response http.Response

	s := strings.Split(filename, ".")
	fileType := s[len(s)-1]
	var contentType string
	switch fileType {
	case "html":
		contentType = "text/html"
	case "txt":
		contentType = "text/plain"
	case "gif":
		contentType = "image/gif"
	case "jpeg":
		contentType = "image/jpeg"
	case "jpg":
		contentType = "image/jpeg"
	case "css":
		contentType = "text/css"
	}
	if contentType == "" {
		response = http.Response{
			Body:       io.NopCloser(bytes.NewBufferString("<h3>400 Bad Request</h3>")),
			Status:     "400 Bad Request",
			StatusCode: 400,
			Proto:      "HTTP/1,1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header, 0),
		}
	} else if err != nil {
		response = http.Response{
			Body:       io.NopCloser(bytes.NewBufferString("<h3>404 Not Found</h3>")),
			Status:     "404 Not Found",
			StatusCode: 404,
			Proto:      "HTTP/1,1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header, 0),
		}
	} else {
		response = http.Response{
			Body:          io.NopCloser(bytes.NewBuffer(file)),
			Status:        "200 OK",
			StatusCode:    200,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			ContentLength: int64(len(file)),
			Header:        make(http.Header, 0),
		}

		response.Header.Add("Content-Type", contentType)
	}

	return response
}

func handlePostRequest(conn net.Conn, requestFields []string) http.Response {
	var response http.Response
	return response
}

func handleInvalidRequest() http.Response {
	var response http.Response
	response = http.Response{
		Body:       io.NopCloser(bytes.NewBufferString("<h3>501 Not Implemented</h3>")),
		Status:     "501 Not Implemented",
		StatusCode: 501,
		Proto:      "HTTP/1,1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header, 0),
	}
	return response
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("[Server] Accepting incoming connection from " + conn.LocalAddr().String())

	s.pool <- conn

	requestFields := extractRequestFields(conn)
	// fmt.Println(requestFields)
	method := requestFields[0]
	var response http.Response
	switch method {
	case "GET":
		response = handleGetRequest(conn, requestFields)
	case "POST":
		response = handlePostRequest(conn, requestFields)
	default:
		response = handleInvalidRequest()
	}

	buff := bytes.NewBuffer(nil)
	response.Write(buff)
	conn.Write(buff.Bytes())

	<-s.pool
}
func (s *Server) Listen() {
	fmt.Println("[Server] Listening for incoming connections...")
	server, err := net.Listen("tcp", ":"+strconv.Itoa(s.port))
	if err != nil {
		fmt.Println("Unable to start server:\n ", err.Error())
	}
	defer server.Close()
	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("Failed to handle client request:\n ", err.Error())
		}
		go s.handleConnection(conn)
	}
}

func main() {
	server := Server{port: 1337, pool: make(chan net.Conn, 10)}
	server.Listen()
}
