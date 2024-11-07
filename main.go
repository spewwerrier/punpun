package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
)

type Server struct {
	address string
	tlspath string
	https   bool
	state   int
}

const (
	SERVER1 = "localhost:4444"
	SERVER2 = "localhost:4445"
	SERVER3 = "localhost:4446"
)

func main() {
	servers := []Server{
		{
			address: SERVER1,
			https:   true,
			state:   1,
		},
		{
			address: SERVER2,
			https:   false,
			state:   1,
		},
		{
			address: SERVER3,
			https:   true,
			state:   1,
		},
	}

	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		panic(err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cer},
	}
	listener, err := tls.Listen("tcp", "localhost:3000", tlsConfig)
	if err != nil {
		panic(err)
	}

	// buffer to store the fields we get from the incoming request
	recvBuf := make([]byte, 1024)
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Printf("accepting a request...\n\n")
		conn.Read(recvBuf)

		b := RoundRobin(recvBuf, servers)
		conn.Write(b)
	}
}

// we dial one of our server with http fields stored in fields buffer
func dial(fields []byte, address string, istls bool) []byte {
	fmt.Printf("dialing %s\n", address)
	recvBuf := make([]byte, 1024)

	if istls {
		cert, err := os.ReadFile("server.crt")
		if err != nil {
			panic(err)
		}
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(cert)

		conf := &tls.Config{
			RootCAs: certPool,
		}

		conn, err := tls.Dial("tcp", address, conf)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		conn.Write(fields)
		_, err = conn.Read(recvBuf)
		if err != nil {
			panic(err)
		}
	} else {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		conn.Write(fields)
		_, err = conn.Read(recvBuf)
		if err != nil {
			panic(err)
		}
	}

	return recvBuf
}

// var servers = []string{
// 	SERVER1,
// 	SERVER2,
// 	SERVER3,
// }

// // we choose any 1 random server and send our dial there
// func RandomRobin(fields []byte) []byte {
// 	i := rand.Int() % len(servers)
// 	return dial(fields, servers[i], true)
// }

// round robin is an arrangement of choosing all elements in a group equally in some rational order, usually from the top to the bottom of a list and then starting again at the top of the list and so on
// this means we have to create states for each server in our load balancer

func (s *Server) IncrementState() {
	s.state += 1
}

func RoundRobin(fields []byte, servers []Server) []byte {
	lowest := servers[0].state

	for _, v := range servers {
		if v.state < lowest {
			lowest = v.state
		}
	}

	// get the server with lowest
	var server string
	istls := false
	for i, v := range servers {
		if v.state == lowest {
			server = v.address
			servers[i].IncrementState()
			istls = servers[i].https
			break
		}
	}

	return dial(fields, server, istls)
}
