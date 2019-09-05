package main

import (
	"fmt"
	"log"
	"net"
)

func connectServer(server string, port int) (net.Conn, error) {
	localAddress, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:%v", *hostAddress, "0"))
	if err != nil {
		log.Fatalf("Error resolving TCP address: %v", err)
	}
	dialer := net.Dialer{
		LocalAddr: localAddress,
	}
	return dialer.Dial("tcp", fmt.Sprintf("%v:%d", server, port))
}
