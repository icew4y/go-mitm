package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	buf := make([]byte, 1024)
	reqBuf := make([]byte, 0)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			log.Printf("Something is wrong while reading data from connection: %v", err)
			break
		}
		log.Printf("Bytes read: %d", n)
		reqBuf = append(reqBuf, buf[0:n]...)
		if strings.Contains(string(reqBuf), "\r\n\r\n") {
			break
		}
	}

	headerEndIndex := bytes.Index(reqBuf, []byte("\r\n\r\n"))
	headerPart := reqBuf[:headerEndIndex+4]
	bodyPart := reqBuf[headerEndIndex+4:]
	request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(headerPart)))
	if err != nil {
		fmt.Printf("http.ReadRequest: %v", err)
		return
	}

	contentLength := 0
	if cl := request.Header.Get("Content-Length"); cl != "" {
		contentLength, err = strconv.Atoi(cl)
		if err != nil {
			fmt.Printf("Invalid Content Length: %s", cl)
			return
		}
	}
	extraBodyLength := contentLength - len(bodyPart)
	fullBody := append([]byte{}, bodyPart...)
	if extraBodyLength > 0 {
		extraBody := make([]byte, extraBodyLength)
		_, err := io.ReadFull(reader, extraBody)
		if err != nil {
			fmt.Printf("Error reading extra body: %v", err)
			return
		}
		fullBody = append(fullBody, extraBody...)
	}

	fmt.Println("Method: ", request.Method)
	fmt.Println("Host: ", request.Host)
	fmt.Println("Path: ", request.URL)
	fmt.Println("Body: ", string(fullBody))
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
	fmt.Printf("Listening on %v", addr)
	for true {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error on accepting connection, abort: %v", err)
		}
		go handleConnection(conn)
	}
}
