package handshake

import (
	"crypto/tls"
	"fmt"
	"log"
)

func RunTlsClient() {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", "127.0.0.1:8443", config)
	if err != nil {
		log.Fatalf("tls.Dial: %v", err)
	}

	defer conn.Close()

	// not needed, tls.Dial already does the handshake
	//err = conn.Handshake()
	//if err != nil {
	//	log.Fatalf("TLS handshake failed: %v", err)
	//}

	fmt.Println("TLS Handshake succeeded!")
	// Read server response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("failed to read from server: %v", err)
	}
	fmt.Println("Server says:", string(buf[:n]))
}
