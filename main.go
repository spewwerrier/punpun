package main

import (
	"fmt"
	"math/rand"
	"net"
)

const (
	SERVER1 = "localhost:4444"
	SERVER2 = "localhost:4445"
	SERVER3 = "localhost:4446"
)

var servers = []string{
	SERVER1,
	SERVER2,
	SERVER3,
}

func main() {
	listener, err := net.Listen("tcp", "localhost:3000")
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

		b := RoundRobin(recvBuf)
		conn.Write(b)
	}
}

// we dial one of our server with http fields stored in fields buffer
func dial(fields []byte, address string) []byte {
	fmt.Printf("dialing %s\n", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	recvBuf := make([]byte, 1024)

	conn.Write(fields)
	_, err = conn.Read(recvBuf)
	if err != nil {
		panic(err)
	}
	return recvBuf
}

// we choose any 1 random server and send our dial there
func RandomRobin(fields []byte) []byte {
	i := rand.Int() % len(servers)
	return dial(fields, servers[i])
}

// round robin is an arrangement of choosing all elements in a group equally in some rational order, usually from the top to the bottom of a list and then starting again at the top of the list and so on
// this means we have to create states for each server in our load balancer
var StatefulServers = map[string]int{
	SERVER1: 0,
	SERVER2: 0,
	SERVER3: 0,
}

func RoundRobin(fields []byte) []byte {
	// we need to get lowest value from our stateful servers
	prev := StatefulServers[SERVER1]

	for _, v := range StatefulServers {
		if v < prev {
			prev = v
		}
	}

	var server string
	for k, v := range StatefulServers {
		if v == prev {
			server = k
		}
	}

	StatefulServers[server]++

	return dial(fields, server)
}
