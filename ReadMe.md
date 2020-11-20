# ftp-port-proxy

Acts as an FTP server that will proxy all requests to an upstream active FTP server and translate the PORT command to any external IP that you choose.

``` sh
Usage of ./ftp-port-proxy:
  -ext-ip string
        The public IP to rewrite FTP port commands from.
  -host-address string
        The IP address to bind the port on. (default "0.0.0.0")
  -host-port int
        The port that this FTP proxy will serve on. (default 20021)
  -server string
        The FTP server host or IP to connect to.
  -server-port int
        The FTP server port number. (default 21)
  -logtostderr=false
        Logs are written to standard error instead of to files.
  -stderrthreshold=ERROR
        Log events at or above this severity are logged to standard
        error as well as to files.
  -log_dir=""
        Log files will be written to this directory instead of the
        default temporary directory.
  -version
        Show version number
```
