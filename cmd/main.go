package main

import (
	"bufio"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net"
	"os"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			log.Printf("Something is wrong while reading data from connection: %v", err)
			break
		}
		log.Printf("Bytes read: %d", n)
		fmt.Printf("Received data: %s\n", buf)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file")
	}
	addr := os.Getenv("LISTEN_ADDR")
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("net.Listen failed to listen on %v", addr)
	}
	log.Printf("Listening on %v", addr)
	for true {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error on accepting connection, abort: %v", err)
		}
		go handleConnection(conn)
	}
}
