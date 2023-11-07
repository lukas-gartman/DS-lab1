package main

import (
	"errors"
	"fmt"
	"net"
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

func handleGetRequest(requestFields []string) {
	path := requestFields[1] // TODO: escape path?
	if _, err := os.Stat(path); err == nil {

	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist

	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence

	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Accepting incoming connection from " + conn.LocalAddr().String())

	s.pool <- conn

	requestFields := extractRequestFields(conn)
	fmt.Println(requestFields)
	method := requestFields[0]
	switch method {
	case "GET":
		fmt.Println("is a GET request")
	case "POST":
		fmt.Println("is a POST request")
	}

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
