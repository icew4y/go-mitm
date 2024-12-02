package go_mitm

import (
	"crypto/tls"
	"fmt"
	"github.com/icew4y/go-mitm/certs"
	"io"
	"log"
	"net"
)

type MITMProxy struct {
	certManager *certs.CertManager
	listenAddr  string
}

func NewMITMProxy(addr string) (*MITMProxy, error) {
	certM, err := certs.NewCertManager()
	if err != nil {
		return nil, fmt.Errorf("certs.NewCertManager: %v", err)
	}
	return &MITMProxy{
		certManager: certM,
		listenAddr:  addr,
	}, nil
}

func (p *MITMProxy) peekClientHello(conn net.Conn) (*tls.ClientHelloInfo, error) {
	peek := make([]byte, 1024)
	n, err := conn.Read(peek)
	if err != nil {
		return nil, err
	}
	clientHello, err := p.parseClientHello(peek[:n])
	if err != nil {
		return nil, err
	}
	return clientHello, nil
}

func (p *MITMProxy) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	clientHello, err := p.peekClientHello(clientConn)
	if err != nil {
		log.Printf("Failed to peek client hello message: %v", err)
		return
	}

	serverName := clientHello.ServerName
	if serverName == "" {
		log.Printf("No SNI header found")
		return
	}

	sniCert, err := p.certManager.GenerateCertificate(serverName)
	if err != nil {
		log.Printf("Failed to GenerateCertificate for %s: %v", serverName, err)
		return
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*sniCert},
	}

	// Upgrade client connection to TLS
	tlsClientConn := tls.Server(clientConn, tlsConfig)
	defer tlsClientConn.Close()

	// Connect to the real server
	realServerConn, err := tls.Dial("tcp", serverName+":443", tlsConfig)
	if err != nil {
		log.Printf("Failed to tls.Dial to server %s:%v", serverName, err)
		return
	}

	defer realServerConn.Close()

	// start bidirectional copy
	go io.Copy(realServerConn, tlsClientConn)
	io.Copy(tlsClientConn, realServerConn)
}

func (p *MITMProxy) parseClientHello(data []byte) (*tls.ClientHelloInfo, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("data too short")
	}

	// Check if it's a TLS handshake
	if data[0] != 0x16 { // Handshake protocol
		return nil, fmt.Errorf("not a handshake")
	}

	// Skip record header (5 bytes)
	// 1 byte  - content type
	// 2 bytes - version
	// 2 bytes - length
	pos := 5

	if len(data) < pos+4 {
		return nil, fmt.Errorf("data too short")
	}

	// Check if it's a ClientHello
	if data[pos] != 0x01 { // ClientHello type
		return nil, fmt.Errorf("not a ClientHello")
	}

	// Skip handshake header
	// 1 byte  - handshake type
	// 3 bytes - length
	pos += 4

	if len(data) < pos+2 {
		return nil, fmt.Errorf("data too short")
	}

	// Skip client version
	pos += 2

	if len(data) < pos+32 {
		return nil, fmt.Errorf("data too short")
	}

	// Skip client random
	pos += 32

	// Skip session ID
	if len(data) < pos+1 {
		return nil, fmt.Errorf("data too short")
	}
	sessionIDLen := int(data[pos])
	pos++
	pos += sessionIDLen

	// Skip cipher suites
	if len(data) < pos+2 {
		return nil, fmt.Errorf("data too short")
	}
	cipherSuitesLen := int(data[pos])<<8 | int(data[pos+1])
	pos += 2
	pos += cipherSuitesLen

	// Skip compression methods
	if len(data) < pos+1 {
		return nil, fmt.Errorf("data too short")
	}
	compressionMethodsLen := int(data[pos])
	pos++
	pos += compressionMethodsLen

	// Check for extensions
	if len(data) < pos+2 {
		return nil, fmt.Errorf("no extensions")
	}
	extensionsLen := int(data[pos])<<8 | int(data[pos+1])
	pos += 2

	// Parse extensions
	endExtensions := pos + extensionsLen
	for pos < endExtensions {
		if len(data) < pos+4 {
			return nil, fmt.Errorf("extension data too short")
		}

		extensionType := uint16(data[pos])<<8 | uint16(data[pos+1])
		extensionLen := int(data[pos+2])<<8 | int(data[pos+3])
		pos += 4

		// Check for SNI extension (type 0)
		if extensionType == 0 {
			// Skip SNI list length
			if len(data) < pos+2 {
				return nil, fmt.Errorf("SNI data too short")
			}
			pos += 2

			// Check SNI type (should be 0 for hostname)
			if len(data) < pos+1 {
				return nil, fmt.Errorf("SNI data too short")
			}
			if data[pos] != 0 {
				return nil, fmt.Errorf("not a hostname SNI")
			}
			pos++

			// Get hostname length and value
			if len(data) < pos+2 {
				return nil, fmt.Errorf("SNI data too short")
			}
			hostnameLen := int(data[pos])<<8 | int(data[pos+1])
			pos += 2

			if len(data) < pos+hostnameLen {
				return nil, fmt.Errorf("SNI data too short")
			}
			hostname := string(data[pos : pos+hostnameLen])

			return &tls.ClientHelloInfo{
				ServerName: hostname,
			}, nil
		}

		pos += extensionLen
	}

	return nil, fmt.Errorf("no SNI extension found")
}

func (p *MITMProxy) Run() error {
	listen, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		return fmt.Errorf("net.Listen: %v", err)
	}
	defer listen.Close()

	fmt.Printf("MITMProxy listening on %s\n", p.listenAddr)
	for {
		clientConn, err := listen.Accept()
		if err != nil {
			fmt.Printf("Failed to accept client connection: %v", err)
			continue
		}
		go func(conn net.Conn) {
			p.handleConnection(conn)
		}(clientConn)
	}
}
