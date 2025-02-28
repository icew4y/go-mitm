package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func copyConnectTunnel(dst io.Writer, src io.Reader, donec chan<- bool) {
	_, err := io.Copy(dst, src)
	if err != nil {
		fmt.Printf("copyConnectTunnel finished with error: %v\n", err)
	}
	fmt.Println("copyConnectTunnel finished")
	donec <- true
}

func handleConnectConnection(conn net.Conn, host string) {
	log.Default().Printf("handleConnectConnection, host: %s\n", host)
	// connect to target server host
	targetConn, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Printf("net.Dial failed to connect to the target server: %s, err: %v", host, err)
		return
	}
	defer targetConn.Close()
	// responses HTTP 200 Connection Established
	_, err = conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		fmt.Printf("conn.Write failed to write response to client: %v\n", err)
		return
	}
	// use io.Copy to forward traffic packets
	donec := make(chan bool, 2)
	go copyConnectTunnel(targetConn, conn, donec)
	go copyConnectTunnel(conn, targetConn, donec)
	// Wait for first copy operation to complete
	log.Default().Println("established CONNECT tunnel, proxying traffic")
	<-donec
	<-donec
	log.Default().Println("Done")
}

func validateProxyAuth(authHeader string) bool {
	if authHeader == "" {
		return false
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Basic" {
		return false
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	creds := strings.SplitN(string(decoded), ":", 2)
	if len(creds) != 2 {
		return false
	}

	prUser := os.Getenv("PROXY_AUTH_USER")
	prPwd := os.Getenv("PROXY_AUTH_PWD")
	return creds[0] == prUser && creds[1] == prPwd
}

func sendProxyAuthRequired(conn net.Conn) {
	_, err := conn.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\n\r\n"))
	if err != nil {
		return
	}
}

func hanldeNormalHttp(req *http.Request, body []byte) (*http.Response, error) {
	newReq, _ := http.NewRequest(req.Method, req.URL.String(), bytes.NewReader(body))
	client := http.Client{
		Timeout: time.Duration(time.Second * 30),
	}
	newReq.Header = req.Header
	resp, err := client.Do(newReq)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

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

	isAuthEnable := os.Getenv("ENABLE_AUTH")
	if isAuthEnable == "true" {
		authHeader := request.Header.Get("Proxy-Authorization")
		if !validateProxyAuth(authHeader) {
			fmt.Printf("Authentication required or failed\n")
			sendProxyAuthRequired(conn)
			return
		}
	}

	if request.Method == "CONNECT" {
		handleConnectConnection(conn, request.Host)
	} else {
		resp, err := hanldeNormalHttp(request, fullBody)
		if err != nil {
			fmt.Printf("hanldeNormalHttp error: %v", err)
			return
		}
		if err = resp.Write(conn); err != nil {
			fmt.Printf("Failed to send back the response: %v", err)
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
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
