package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type dataProxy struct {
	rconn net.Conn
	l     net.Listener
	close chan error
}

func (d *dataProxy) start(originalPort, newPort int, originalIP string) error {
	// Listen for connections from upstream server on new port
	l, err := net.Listen("tcp", fmt.Sprintf("%v:%v", hostAddress, newPort))
	if err != nil {
		return err
	}
	d.l = l

	// Create connection back to source server on original port
	c, err := connectServer(originalIP, originalPort)
	if err != nil {
		return err
	}
	d.rconn = c
	log.Println("Connected reverse proxy to", originalIP, originalPort)

	d.close = make(chan error, 2)

	return nil
}

func (d *dataProxy) serve(ready chan bool) {
	log.Println("Waiting for data on", d.l.Addr().String())
	for {
		ready <- true
		c, err := d.l.Accept()
		if err != nil {
			log.Printf("Error accepting data connection: %v", err)
			return
		}
		go d.handleConnection(c)
	}
}

func (d *dataProxy) handleConnection(c net.Conn) {
	defer c.Close()
	defer d.rconn.Close()
	defer d.l.Close()
	log.Printf("Accepted data connection from %v", c.RemoteAddr())

	// Bidirectional proxy
	go d.proxyData(d.rconn, c)
	go d.proxyData(c, d.rconn)
	<-d.close

	log.Printf("Closed connection from %v", c.RemoteAddr())
}

func (d *dataProxy) err(err error) {
	if err != io.EOF {
		log.Printf("Error: %v", err)
	}
	d.close <- err
}

func (d *dataProxy) proxyData(lconn, rconn io.ReadWriter) {
	buff := make([]byte, 4096)
	for {
		n, err := lconn.Read(buff)
		if err != nil {
			d.err(err)
			return
		}
		b := buff[:n]

		n, err = rconn.Write(b)
		if err != nil {
			d.err(err)
			return
		}
		log.Printf("Copied %d data bytes", n)
	}
}

type ftpProxy struct {
	UpstreamServer string
	UpstreamPort   int
	errsig         chan error
	lconn          io.ReadWriter
	rconn          io.ReadWriter
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

		log.Printf("Wrote: %s", b)

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
		replacementCmd, newPort, originalPort, originalIP := translatePortCommand(b)
		log.Println("Going to listen on port", newPort, "and proxy connection back to", originalIP, "on port", originalPort)

		// Start data proxy
		d := &dataProxy{}
		err = d.start(originalPort, newPort, originalIP)
		if err != nil {
			log.Println("Unable to start data listening connection on port", newPort)
			return
		}
		readyToServe := make(chan bool)
		go d.serve(readyToServe)
		<-readyToServe

		log.Printf("Translated to: %v", strings.Trim(string(replacementCmd), "\n"))
		return w.Write(replacementCmd)
	}
	return w.Write(b)
}

func (f *ftpProxy) handleConnection(lc net.Conn) {
	log.Printf("Accepted control connection from %v", lc.RemoteAddr())

	// Start connection to upstream server
	uc, err := connectServer(f.UpstreamServer, f.UpstreamPort)
	if err != nil {
		log.Printf("Failed to connect to upstream %v", f.UpstreamServer)
		return
	}

	// Bidirectional proxy
	f.lconn = lc
	f.rconn = uc
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
