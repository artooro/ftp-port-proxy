package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func serveFTP(port int) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("0.0.0.0:%v", port))
}

func translatePortCommand(cmd []byte) []byte {
	// PORT 127,0,0,1,234,229
	extIP := strings.Split(*externalIP, ".")
	ipPart := strings.Join(extIP, ",")
	data := strings.Split(string(cmd), ",")
	return []byte(fmt.Sprintf("PORT %v,%v,%v", ipPart, data[4], data[5]))
}

func parsePort(command string) int {
	data := strings.Split(strings.Trim(command, "\n\r"), ",")
	if len(data) < 5 {
		return 0
	}
	p1, err := strconv.Atoi(data[4])
	if err != nil {
		return 0
	}
	p2, err := strconv.Atoi(data[5])
	if err != nil {
		return 0
	}
	port := p1*256 + p2

	return port
}
