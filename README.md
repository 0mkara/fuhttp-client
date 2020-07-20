# FuHTTP Client
This client mocks browser configurations and perform requests to given URL using uTLS configured connections.

`fuhttp-client` uses `fuhttp` a forked module of (fasthttp)[https://github.com/valyala/fasthttp] to perform TLS handshake.

This client will create an unix socket server and listen for raw requests. Upon request message it will process the HTTP/HTTPS request with provided configuration and emit results. It can also retain sessions.

# Development
```
go clean --modcache
go mod tidy
go run *.go
```

# Build & run
* Linux
```
go build
./client
```
* Windows
```
GOOS=windows GOARCH=amd64 go build
client.exe
```