# go-mitm
A simple mitm proxy written in Go


## How a MITM Proxy Works
A MITM proxy intercepts and manipulates network traffic between a client and a server. It acts as an intermediary that decrypts, inspects, and potentially alters HTTP/HTTPS traffic.


## 1.1 Components of a MITOM Proxy
1. Listener - Captures incoming connections
2. TLS Interception (for HTTPS) - Uses a self-signed CA certificate to decrypt encrypted traffic
3. Request handling - Parses, modifies, or logs HTTP requests
4. Response handling - Parses, modifies, or logs HTTP responses
5. Forwarding Mechanism - Sends traffic to the original destination or an alternate one
6. Logging & Debugging - Stores intercepted traffic for analysis


## Step1: Set Up a Basic HTTP Proxy
- Implement an HTTP proxy that listens on a port
- Capture and forward HTTP requests and responses

## Step2: Add TLS Support for HTTPS Interception
- Generate a self-signed CA certificate
- Sign leaf certificates dynamically for different domains
- Use `crypto/tls` and `mitmproxy` techniques to decrypt traffic

## Step3: Parse and Modify Requests and Responses
- Decode HTTP headers, body, and parameters
- Allow modifications (e.g., inject headers, replace payloads)
- Support traffic inspection for debugging

## Step4: Implement Forwarding Logic
- Set up a TCP tunneling mechanism for HTTPS
- May use golang built-in support `io.Copy`

## Step5: Logging and Debugging
- Store HTTP request/response details in database
- Implement real-time monitoring via WebSocket or API

## Step6: Support Custom Rules and Filters
- Allow users to define rules (e.g. blocking by domains, modify content)

## Step7: Improve Performance and Security
- Use goroutines go handle concurrent requests
- Optimize for **high throughput** with connection pooling



# Components need to be implemented

### Proxy Server
- Listen on a port and handles HTTP/HTTPS requests
- Use golang built-in `net/http` or `net` packages for networking

### TLS Interceptor
- Uses `crypto/tls` to decrypt HTTPS traffic
- Dynamically generates per-site certificate

### Request and Response Parser
- TODO: To be added

### Forwarding
- TODO: To be added

### Logging system
- TODO: To be added

### Configuration & Rule Engine
- TODO: To be added

