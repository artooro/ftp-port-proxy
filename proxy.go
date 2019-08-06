package main

import (
	"io"
	"log"
	"net"
	"strings"
)

type ftpProxy struct {
	UpstreamServer string
	UpstreamPort   int
	errsig         chan error
}

func (f *ftpProxy) proxyData(lconn, rconn io.ReadWriter) {
	buff := make([]byte, 4096)
	for {
		n, err := lconn.Read(buff)
		if err != nil {
			f.err(err)
			return
		}
		b := buff[:n]

		// The port command translation will happen here
		log.Printf("Wrote: %s", b)

		//n, err = rconn.Write(b)
		n, err = f.translator(b, rconn)
		if err != nil {
			f.err(err)
			return
		}
		log.Printf("Copied %d bytes", n)
	}
}

func (f *ftpProxy) translator(b []byte, w io.ReadWriter) (n int, err error) {
	portCmd := []byte("PORT")
	if len(b) < len(portCmd) {
		return w.Write(b)
	}
	cmdStr := string(b[:len(portCmd)])
	switch cmdStr {
	case "PORT":
		replacementCmd := translatePortCommand(b)
		portInt := parsePort(string(b))
		log.Printf("Translated to: %v", strings.Trim(string(replacementCmd), "\n"))
		log.Println("FTP data port:", portInt)
		w.Write(replacementCmd)
		return
	}
	return w.Write(b)
}

func (f *ftpProxy) handleConnection(lc net.Conn) {
	log.Printf("Accepted connection from %v", lc.RemoteAddr())

	// Start connection to upstream server
	uc, err := connectServer(f.UpstreamServer, f.UpstreamPort)
	if err != nil {
		log.Printf("Failed to connect to upstream %v", f.UpstreamServer)
		return
	}

	// Bidirectional proxy
	go f.proxyData(lc, uc)
	go f.proxyData(uc, lc)

	<-f.errsig
	log.Printf("Closed connection %v <> %v", lc.RemoteAddr(), uc.RemoteAddr())
}

func (f *ftpProxy) err(err error) {
	if err != io.EOF {
		log.Printf("Error: %v", err)
	}
	f.errsig <- err
}

func (f *ftpProxy) listenAndServe(l net.Listener) error {
	f.errsig = make(chan error, 2)

	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go f.handleConnection(c)
	}
}
