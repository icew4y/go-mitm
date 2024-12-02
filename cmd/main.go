package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	go_mitm "github.com/icew4y/go-mitm"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

type fakeResponseWriter struct {
	conn   net.Conn
	header http.Header
	status int
}

func (w *fakeResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *fakeResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	if w.status == 0 {
		w.status = http.StatusOK
	}

	// Write the HTTP status line
	_, err := w.conn.Write([]byte("HTTP/1.1 " + http.StatusText(w.status) + "\r\n"))
	if err != nil {
		return
	}

	// Write the headers
	for key, values := range w.Header() {
		for _, value := range values {
			_, _ = w.conn.Write([]byte(key + ": " + value + "\r\n"))
		}
	}
	// End of headers
	_, _ = w.conn.Write([]byte("\r\n"))
}

func (w *fakeResponseWriter) Write(data []byte) (int, error) {
	return w.conn.Write(data)
}

func loadCertificate() tls.Certificate {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatalf("Failed to load certificate: %v", err)
	}
	return cert
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{
		Timeout: time.Duration(time.Second * 30),
	}

	newReq, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		log.Printf("http.NewRequest: %v", err)
		http.Error(w, "http.NewRequest", http.StatusInternalServerError)
		return
	}

	newReq.Header = r.Header

	resp, err := client.Do(newReq)
	if err != nil {
		log.Printf("client.Do: %v", err)
		http.Error(w, "client.Do", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for key, val := range resp.Header {
		for _, v := range val {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("io.Copy: %v", err)
	}
}

func handleHTTPS(conn net.Conn, host string) {
	// Establish a connection to the destination
	destConn, err := net.DialTimeout("tcp", host, 10*time.Second)
	if err != nil {
		log.Printf("Failed to connect to destination: %v", err)
		conn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		conn.Close()
		return
	}
	defer destConn.Close()

	// Respond to the client with 200 OK to start the tunnel
	conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	// Forward traffic between client and destination
	go io.Copy(destConn, conn)
	io.Copy(conn, destConn)
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	// Read the first request
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("Failed to read from connection: %v", err)
		return
	}

	// Parse the request
	requestLine := string(buf[:n])
	if len(requestLine) == 0 {
		return
	}

	var method, host string
	_, err = fmt.Sscanf(requestLine, "%s%s", &method, &host)
	if err != nil {
		log.Printf("Failed to parse request line: %v", err)
		return
	}

	if method == "CONNECT" {
		handleHTTPS(conn, host)
	} else {
		// Fallback to HTTP handling
		w := &fakeResponseWriter{conn: conn}
		r, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(buf[:n])))
		if err != nil {
			log.Printf("Failed to parse HTTP request: %v", err)
			return
		}
		handleHTTP(w, r)
	}
}

func main() {
	//listener, err := net.Listen("tcp", ":8080")
	//if err != nil {
	//	log.Fatalf("Failed to start listener: %v", err)
	//}
	//defer listener.Close()
	//
	//log.Println("Starting Proxy on :8080")
	//
	//for {
	//	conn, err := listener.Accept()
	//	if err != nil {
	//		log.Printf("Failed to accept connection: %v", err)
	//		continue
	//	}
	//	go handleRequest(conn)
	//}

	mitm, err := go_mitm.NewMITMProxy("0.0.0.0:8443")
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}
	err = mitm.Run()
	if err != nil {
		log.Fatalf("Failed to mitm.Run: %v", err)
	}
}
