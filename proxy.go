package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/golang/glog"
)

type dataProxy struct {
	rconn net.Conn
	l     net.Listener
	close chan error
}

func (d *dataProxy) start(originalPort, newPort int, originalIP string) error {
	// Listen for connections from upstream server on new port
	l, err := net.Listen("tcp", fmt.Sprintf("%v:%v", *hostAddress, newPort))
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
	glog.Infoln("Connected reverse proxy to", originalIP, originalPort)

	d.close = make(chan error, 2)

	return nil
}

func (d *dataProxy) serve() {
	glog.Infoln("Waiting for data on", d.l.Addr().String())

	c, err := d.l.Accept()
	if err != nil {
		glog.Infof("Error accepting data connection: %v", err)
		return
	}
	go d.handleConnection(c)

}

func (d *dataProxy) handleConnection(c net.Conn) {
	defer c.Close()
	defer d.rconn.Close()
	defer d.l.Close()
	glog.Infof("Accepted data connection from %v", c.RemoteAddr())

	// Bidirectional proxy
	go d.proxyData(d.rconn, c)
	go d.proxyData(c, d.rconn)
	<-d.close

	glog.Infof("Closed connection from %v", c.RemoteAddr())
}

func (d *dataProxy) err(err error) {
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
		glog.Infof("Copied %d data bytes", n)
	}

}

type ftpProxy struct {
	UpstreamServer string
	UpstreamPort   int
	errsig         chan error
	lconn          io.ReadWriter
	rconn          io.ReadWriter
}

func (f *ftpProxy) Execute(signals chan os.Signal) {
	l, err := serveFTP(*port)
	if err != nil {
		glog.Fatalf("Unable to listen on port %v (err: %v)", *port, err)
	}

	f.UpstreamServer = *server
	f.UpstreamPort = *serverPort

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go f.listenAndServe(ctx, l)

	for {
		select {
		case <-signals:
			cancel()
			os.Exit(1)
		}
	}
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

		glog.Infof("Wrote: %s", b)

		n, err = f.translator(b, rconn)
		if err != nil {
			f.err(err)
			return
		}
		glog.Infof("Copied %d bytes", n)
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
		glog.Infof("Going to listen on port", newPort, "and proxy connection back to", originalIP, "on port", originalPort)

		// Start data proxy
		d := &dataProxy{}
		err = d.start(originalPort, newPort, originalIP)
		if err != nil {
			glog.Warningf("Unable to start data listening connection on port", newPort)
			return
		}

		go d.serve()

		glog.Infof("Translated to: %v", strings.Trim(string(replacementCmd), "\n"))
		return w.Write(replacementCmd)
	}
	return w.Write(b)
}

func (f *ftpProxy) handleConnection(lc, uc net.Conn) {
	glog.Infof("Accepted control connection from %v", lc.RemoteAddr())

	// Bidirectional proxy
	go f.proxyData(lc, uc)
	go f.proxyData(uc, lc)

	<-f.errsig
	glog.Warningf("Closed connection %v <> %v", lc.RemoteAddr(), uc.RemoteAddr())
}

func (f *ftpProxy) err(err error) {
	if err != io.EOF {
		glog.Infof("Error: %v", err)
	}
	f.errsig <- err
}

func (f *ftpProxy) listenAndServe(ctx context.Context, l net.Listener) error {
	f.errsig = make(chan error, 2)

	for {
		select {
		case <-ctx.Done():
			l.Close()
			return nil
		default:
			c, err := l.Accept()
			if err != nil {
				glog.Errorf("Error accepting connection: %v", err)
				continue
			}

			// Start connection to upstream server
			uc, err := connectServer(f.UpstreamServer, f.UpstreamPort)
			if err != nil {
				glog.Errorf("Failed to connect to upstream %v", f.UpstreamServer)
				continue
			}

			go f.handleConnection(c, uc)
		}
	}
}
