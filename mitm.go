package go_mitm

import (
	"net"
	"net/http"
)

// https://medium.com/@andrea.grillo96/implementing-an-http-proxy-in-go-50c8d6a24985
type Proxy struct {
	addr net.Addr

	transport http.RoundTripper

	listener net.Listener
}
