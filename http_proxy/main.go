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

type Proxy struct {
	proxyPort     int
	serverAddress string
	pool          chan net.Conn
}

func (p *Proxy) handleGetRequest(conn net.Conn, request *http.Request) http.Response {
	var requestURI string
	port, _ := strconv.Atoi(strings.Split(request.Host, ":")[1])
	if port == p.proxyPort { // Request directly to proxy (e.g. firefox)
		requestURI = "http://" + p.serverAddress + request.RequestURI
	} else { // Request via proxy (e.g. curl)
		requestURI = request.RequestURI
	}

	response, err := http.Get(requestURI)
	if err != nil {
		fmt.Println(err.Error())
		return *response
	}

	fmt.Println("[Proxy] Received response from server")
	return *response
}

func (p *Proxy) handleInvalidRequest() http.Response {
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

func (p *Proxy) handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("[Proxy] Accepting incoming connection from " + conn.RemoteAddr().String())

	p.pool <- conn

	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}

	method := request.Method
	var response http.Response
	switch method {
	case "GET":
		response = p.handleGetRequest(conn, request)
	default:
		response = p.handleInvalidRequest()
	}

	buff := bytes.NewBuffer(nil)
	response.Write(buff)
	conn.Write(buff.Bytes())

	<-p.pool
}

func (p *Proxy) Listen() {
	proxy, err := net.Listen("tcp", ":"+strconv.Itoa(p.proxyPort))
	if err != nil {
		fmt.Println("Unable to start proxy:\n ", err.Error())
	}
	defer proxy.Close()

	fmt.Println("[Proxy] Listening for incoming connections on port " + strconv.Itoa(p.proxyPort) + "...")
	for {
		conn, err := proxy.Accept()
		if err != nil {
			fmt.Println("Failed to handle request:\n ", err.Error())
		}
		go p.handleConnection(conn)
	}
}

func main() {
	if len(os.Args) > 2 {
		serverInput := strings.Split(os.Args[2], ":")
		if len(serverInput) != 2 {
			fmt.Println("Please provide a proxy port and a server address with a valid port (ex: 1337 127.0.0.1:8080).")
			os.Exit(-1)
		}
		proxyPortInt, proxyErr := strconv.Atoi(os.Args[1])
		serverPortInt, serverErr := strconv.Atoi(serverInput[1])
		if proxyErr != nil || serverErr != nil || proxyPortInt < 1024 || proxyPortInt > 65353 || serverPortInt < 1024 || serverPortInt > 65353 {
			fmt.Println("Invalid ports. Enter ports within range 1024-65353.")
			os.Exit(-1)
		}

		proxy := Proxy{proxyPort: proxyPortInt, serverAddress: os.Args[2], pool: make(chan net.Conn, 10)}
		proxy.Listen()
	} else {
		fmt.Println("Program must be started with a proxy port and server address as arguments (ex: 127.0.0.1:1337 127.0.0.1:8080).")
	}
}
