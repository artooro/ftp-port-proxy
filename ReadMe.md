# ftp-port-proxy

Acts as an FTP server that will proxy all requests to an upstream active FTP server and translate the PORT command to any external IP that you choose.

``` sh
./ftp-port-proxy -help
  -ext-ip string
        The public IP to rewrite FTP port commands from.
  -host-port int
        The port that this FTP proxy will serve on. Defaults to 20021. (default 20021)
  -server string
        The FTP server host or IP to connect to.
  -server-port int
        The FTP server port number. Defaults to 21. (default 21)
```
