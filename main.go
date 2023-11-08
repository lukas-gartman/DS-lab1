package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Server struct {
	port int
	pool chan net.Conn
}

func getContentType(filename string) (string, bool) {
	var contentType string
	error := false

	s := strings.Split(filename, ".")
	fileType := s[len(s)-1]
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
	default:
		error = true
	}

	return contentType, error
}

func getFileType(contentType string) (string, bool) {
	var fileType string
	error := false

	s := strings.Split(contentType, ";")
	ct := strings.TrimSpace(strings.ToLower(s[0]))
	switch ct {
	case "text/html":
		fileType = ".html"
	case "text/plain":
		fileType = ".txt"
	case "image/gif":
		fileType = ".gif"
	case "image/jpeg":
		fileType = ".jpg"
	case "text/css":
		fileType = ".css"
	default:
		error = true
	}

	return fileType, error
}

func handleGetRequest(conn net.Conn, request *http.Request) http.Response {
	pwd, _ := os.Getwd()
	filename := request.RequestURI
	fmt.Println(filename)
	file, fileError := os.ReadFile(pwd + filename)
	var response http.Response
	contentType, contentTypeError := getContentType(filename)

	if contentTypeError {
		response = http.Response{
			Body:       io.NopCloser(bytes.NewBufferString("<h3>400 Bad Request</h3>")),
			Status:     "400 Bad Request",
			StatusCode: 400,
			Proto:      "HTTP/2",
			ProtoMajor: 2,
			ProtoMinor: 0,
			Header:     make(http.Header, 0),
		}
	} else if fileError != nil {
		response = http.Response{
			Body:       io.NopCloser(bytes.NewBufferString("<h3>404 Not Found</h3>")),
			Status:     "404 Not Found",
			StatusCode: 404,
			Proto:      "HTTP/2",
			ProtoMajor: 2,
			ProtoMinor: 0,
			Header:     make(http.Header, 0),
		}
	} else {
		response = http.Response{
			Body:          io.NopCloser(bytes.NewBuffer(file)),
			Status:        "200 OK",
			StatusCode:    200,
			Proto:         "HTTP/2",
			ProtoMajor:    2,
			ProtoMinor:    0,
			ContentLength: int64(len(file)),
			Header:        make(http.Header, 0),
		}
		response.Header.Add("Content-Type", contentType)
	}

	return response
}

func readData(conn net.Conn, bufferSize int) []byte {
	var data []byte
	for {
		buffer := make([]byte, bufferSize)
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				data = append(data, buffer[:n]...)
				break
			} else {
				fmt.Println("Error reading:\n ", err.Error())
			}
		}
		data = append(data, buffer[:n]...)
	}

	return data
}

func handlePostRequest(conn net.Conn, request *http.Request) http.Response {
	var response http.Response
	// filename := request.RequestURI
	contentType := request.Header.Get("Content-Type")
	fileType, fileTypeError := getFileType(contentType)
	if fileTypeError {
		response = http.Response{
			Body:       io.NopCloser(bytes.NewBufferString("<h3>400 Bad Request</h3>")),
			Status:     "400 Bad Request",
			StatusCode: 400,
			Proto:      "HTTP/2",
			ProtoMajor: 2,
			ProtoMinor: 0,
			Header:     make(http.Header, 0),
		}
	} else {
		data, dataErr := io.ReadAll(request.Body)
		if dataErr != nil {
			fmt.Println(dataErr.Error())
		}
		fileName := request.Header.Get("filename")
		fmt.Println("filename is ", fileName)
		fmt.Println("filetype is ", fileType)
		pwd, _ := os.Getwd()
		fileErr := os.WriteFile(pwd+"/"+fileName+"."+fileType, data, 0644)
		if fileErr != nil {
			fmt.Println(fileErr.Error())
		}
	}
	return response
}

func handleInvalidRequest() http.Response {
	return http.Response{
		Body:       io.NopCloser(bytes.NewBufferString("<h3>501 Not Implemented</h3>")),
		Status:     "501 Not Implemented",
		StatusCode: 501,
		Proto:      "HTTP/2",
		ProtoMajor: 2,
		ProtoMinor: 0,
		Header:     make(http.Header, 0),
	}
}

func extractRequestFields(conn net.Conn) []string {
	// data = readData(conn, 1024)
	buffer := make([]byte, 1024)
	mLen, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading:\n ", err.Error())
	}

	return strings.Fields(string(buffer[:mLen]))
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("[Server] Accepting incoming connection from " + conn.LocalAddr().String())

	s.pool <- conn

	// requestFields := extractRequestFields(conn)
	// fmt.Println(requestFields)
	// method := requestFields[0]

	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}

	method := request.Method
	var response http.Response
	switch method {
	case "GET":
		response = handleGetRequest(conn, request)
	case "POST":
		response = handlePostRequest(conn, request)
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
	if len(os.Args) > 1 {
		port_arg, err := strconv.Atoi(os.Args[1])
		if err == nil {
			server := Server{port: port_arg, pool: make(chan net.Conn, 10)}
			server.Listen()
		}
	} else {
		fmt.Println("Please provide a port.")
	}
}
