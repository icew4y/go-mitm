package handshake

import (
	"crypto/tls"
	"fmt"
	"log"
)

func RunTlsServer() {
	cert, err := tls.LoadX509KeyPair("./certs/cert.pem", "./certs/key.pem")
	if err != nil {
		log.Fatalf("Failed to load cert & key: %v", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", ":8443", config)
	if err != nil {
		log.Fatalf("net.Listen: %v", err)
	}

	defer listener.Close()

	fmt.Println("Server is listening on port 8443.")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		tlsConn := conn.(*tls.Conn)
		err = tlsConn.Handshake()
		if err != nil {
			log.Printf("TLS handshake failed: %v", err)
			tlsConn.Close()
			continue
		}

		fmt.Println("TLS handshake succeeded with:", tlsConn.RemoteAddr())
		go func(c *tls.Conn) {
			defer c.Close()
			c.Write([]byte("Hello from TLS server after handshake!\\n"))
		}(tlsConn)
	}
}
