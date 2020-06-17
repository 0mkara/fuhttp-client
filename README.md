# Ghost FuHTTP Client
This client mocks browser configurations and makes request to given address using provided configurations

# Development
```
go clean --modcache
go mod tidy
go run *.go
```
This will create a unix socket server and listen for requests. Upon request message it will process the HTTP/HTTPS request with provided configuration and emit results.

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