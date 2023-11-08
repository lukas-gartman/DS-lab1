package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
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

func handlePostRequest(conn net.Conn, request *http.Request) http.Response {
	var response http.Response
	uri := request.RequestURI
	fmt.Println(uri)
	contentType := request.Header.Get("Content-Type")
	filename := request.Header.Get("filename")
	fileType, fileTypeError := getFileType(contentType)
	if filename == "" {
		var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
		filename = strconv.Itoa(seededRand.Int())
	}

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

		pwd, _ := os.Getwd()
		dirErr := os.MkdirAll(pwd+uri, os.ModePerm)
		if dirErr != nil {
			fmt.Println(dirErr.Error())
			return http.Response{
				Body:       io.NopCloser(bytes.NewBufferString("<h3>400 Bad Request</h3>")),
				Status:     "400 Bad Request",
				StatusCode: 400,
				Proto:      "HTTP/2",
				ProtoMajor: 2,
				ProtoMinor: 0,
				Header:     make(http.Header, 0),
			}
		} else {
			pwd = pwd + uri
		}
		fileErr := os.WriteFile(pwd+"/"+filename+fileType, data, os.ModePerm)
		if fileErr != nil {
			fmt.Println(fileErr.Error())
		}

		response = http.Response{
			Body:       io.NopCloser(bytes.NewBufferString(uri + "/" + filename + fileType)),
			Status:     "200 OK",
			StatusCode: 200,
			Proto:      "HTTP/2",
			ProtoMajor: 2,
			ProtoMinor: 0,
			Header:     make(http.Header, 0),
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

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("[Server] Accepting incoming connection from " + conn.LocalAddr().String())

	s.pool <- conn

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
	server, err := net.Listen("tcp", ":"+strconv.Itoa(s.port))
	if err != nil {
		fmt.Println("Unable to start server:\n ", err.Error())
	}
	defer server.Close()

	fmt.Println("[Server] Listening for incoming connections on port " + strconv.Itoa(s.port) + "...")
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
