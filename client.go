package main

import (
	"fmt"
	"log"
	"net"
)

func connectServer(server string, port int) (net.Conn, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:%d", server, port))
	if err != nil {
		log.Fatalf("Error resolving TCP address: %v", err)
	}
	return net.DialTCP("tcp", nil, addr)
}
