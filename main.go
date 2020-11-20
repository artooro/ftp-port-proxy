package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const versionString = "0.1.4"

var (
	port        = flag.Int("host-port", 20021, "The port that this FTP proxy will serve on.")
	hostAddress = flag.String("host-address", "0.0.0.0", "The IP address to bind the port on.")
	server      = flag.String("server", "", "The FTP server host or IP to connect to.")
	serverPort  = flag.Int("server-port", 21, "The FTP server port number.")
	externalIP  = flag.String("ext-ip", "", "The public IP to rewrite FTP port commands from.")
	showVersion = flag.Bool("version", false, "Show version number")
)

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Println("ftp-port-proxy:", versionString)
		return
	}

	s := &ftpProxy{}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	s.Execute(c)
}
