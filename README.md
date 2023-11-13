# DS-lab1

Lab 1 of the course Distributed Systems, 2023.

### HTTP server

A simple HTTP server written in Go, implemented without using `http` server functions (e.g. `ListenAndServe`). It supports HTTP methods `GET` and `POST`.

## Usage
 
Starting the server:
```
$ go run main.go <port>
```

Requests can be made through `curl`, Postman, web browsers, etc.

## Methods

`func (s *Server) Listen(conn net.Conn)`
: Creates a listener to the network, that listens to TCP connections to the server's port. Any incoming connections get accepted and handled by the connection handler.

`func (s *Server) handleConnection(conn net.Conn)`
: Assigns workers from the server's pool to handle requests. GET, POST and invalid requests are handled in separate methods.

### HTTP proxy

A simple HTTP server written in Go, supporting proxy connections. It supports HTTP `GET`.

## Usage

Starting the server:
```
$ go run main.go <proxy>:<port> <server>:<port>
```

Requests can be made through `curl`, Postman, web browsers, etc.

## Methods

`func (p *Proxy) Listen(conn net.Conn)`
: Creates a listener to the network, that listens to TCP connections to the proxy server's port. Any incoming connections get accepted and handled by the connection handler.

`func (p *Proxy) handleConnection(conn net.Conn)`
: Assigns workers from the proxy server's pool to handle requests. GET and invalid requests are handled in separate methods. 

`func (p *Proxy) handleGetRequest(conn net.Conn, request *http.Request) http.Response`
: Requests are forwarded to server address.
