package main

import (
	"log"
	"net"
	"os"
	"time"

	tls "github.com/refraction-networking/utls"
	"github.com/valyala/fasthttp"
)

var (
	sockAddr = "/tmp/fuhttp.sock"
	client   = &fasthttp.Client{
		Name:                          "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:69.0) Gecko/20100101 Firefox/69.0",
		NoDefaultUserAgentHeader:      true,
		MaxConnsPerHost:               10000,
		ReadBufferSize:                4 * 4096, // Make sure to set this big enough that your whole request can be read at once.
		WriteBufferSize:               4 * 4096, // Same but for your response.
		ReadTimeout:                   time.Second,
		WriteTimeout:                  time.Second,
		MaxIdleConnDuration:           time.Minute,
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this.
		TLSConfig:                     &tls.Config{InsecureSkipVerify: true},
		ClientHelloID:                 &tls.HelloFirefox_65,
	}
)

func main() {
	log.Println("Starting new IPC server...")
	// Start IPC server
	if err := os.RemoveAll(sockAddr); err != nil {
		log.Fatal(err)
	}
	s, err := net.Listen("unix", sockAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}
	defer s.Close()
	for {
		// Accept new connections, dispatching them to echoServer
		// in a goroutine.
		conn, err := s.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		go echoServer(conn)
		go reader(conn)
	}
}

func echoServer(c net.Conn) {
	log.Printf("Client connected [%s]", c.RemoteAddr().Network())
}
