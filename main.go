package main

import (
	"flag"
	"log"
)

var (
	port        = flag.Int("host-port", 20021, "The port that this FTP proxy will serve on.")
	hostAddress = flag.String("host-address", "0.0.0.0", "The IP address to bind the port on.")
	server      = flag.String("server", "", "The FTP server host or IP to connect to.")
	serverPort  = flag.Int("server-port", 21, "The FTP server port number.")
	externalIP  = flag.String("ext-ip", "", "The public IP to rewrite FTP port commands from.")
)

func main() {
	flag.Parse()

	// Start internal server
	l, err := serveFTP(*port)
	if err != nil {
		log.Fatalf("Unable to listen on port %d (err: %v)", *port, err)
	}

	log.Printf("Serving on port %d", *port)

	proxy := ftpProxy{
		UpstreamServer: *server,
		UpstreamPort:   *serverPort,
	}
	log.Fatal(proxy.listenAndServe(l))
}
