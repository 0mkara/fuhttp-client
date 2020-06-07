package main

import (
	"log"
	"net"
	"os"
)

var (
	sockAddr = "/tmp/fuhttp.sock"
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
