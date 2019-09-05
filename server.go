package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func serveFTP(port int) (net.Listener, error) {
	return net.Listen("tcp", fmt.Sprintf("%v:%v", *hostAddress, port))
}

func translatePortCommand(cmd []byte) (command []byte, newPort, originalPort int, originalIP string) {
	// PORT 127,0,0,1,234,229
	extIP := strings.Split(*externalIP, ".")
	ipPart := strings.Join(extIP, ",")
	originalPort = parsePort(string(cmd))
	newPort = originalPort + 1
	portSection := convertPort(newPort)
	command = []byte(fmt.Sprintf("PORT %v,%v\r\n", ipPart, portSection))
	data := strings.Split(string(cmd), ",")
	originalIP = fmt.Sprintf("%v.%v.%v.%v", strings.Split(data[0], " ")[1], data[1], data[2], data[3])
	return
}

func convertPort(port int) string {
	p1 := port / 256
	p2 := port - (p1 * 256)

	return fmt.Sprintf("%v,%v", p1, p2)
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
