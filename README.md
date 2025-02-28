# go-mitm

A powerful and efficient Man-in-the-Middle (MITM) proxy written in Go, designed for intercepting, analyzing, and modifying HTTP/HTTPS traffic.

## Features

- 🔒 Full HTTPS support with dynamic certificate generation
- 🚀 High-performance concurrent request handling
- 📝 Detailed traffic logging and analysis
- 🛠 Customizable request/response modification
- 🔍 Real-time traffic inspection
- 🎯 Rule-based traffic filtering
- 💻 Cross-platform support (Windows, Linux, macOS)

## Prerequisites

- Go 1.23.2 or higher
- OpenSSL (for certificate generation)

## Installation

```bash
# Clone the repository
git clone https://github.com/icew4y/go-mitm.git

# Change to project directory
cd go-mitm

# Install dependencies
go mod download
```

## Quick Start

1. Generate a CA certificate (first-time setup):
```bash
# Instructions for generating certificates will be provided
```

2. Run the proxy:
```bash
go run cmd/main.go
```

3. Configure your system/browser to use the proxy:
   - Default proxy address: `localhost:8080`
   - Import the generated CA certificate into your browser/system trust store

## Configuration

The proxy can be configured using environment variables or a `.env` file:

```env
PROXY_PORT=8080
# Add other configuration options
```

## Architecture

### Core Components

1. **Proxy Server**
   - Handles incoming connections
   - Supports both HTTP and HTTPS protocols
   - Manages connection pooling for optimal performance

2. **TLS Interceptor**
   - Generates dynamic certificates for HTTPS interception
   - Handles TLS handshakes and protocol negotiation
   - Supports modern TLS features including 0-RTT

3. **Request/Response Handler**
   - Parses and modifies HTTP headers and bodies
   - Supports content encoding/decoding
   - Implements traffic manipulation rules

4. **Logging System**
   - Captures detailed traffic information
   - Supports various output formats
   - Enables real-time monitoring

### Flow Diagram

```
Client <-> TLS Interceptor <-> Request Handler <-> Response Handler <-> Target Server
```


## Step1: Set Up a Basic HTTP Proxy
- [x] Implement an HTTP proxy that listens on a port
- [x] Capture and forward HTTP requests and responses
- [x] HTTP Tunneling for `CONNECT` method
- [x] Proxy Authentication
  ```
  Client                 Proxy                     Server
  |                       |                          |
  | HTTP Request          |                          |
  |---------------------->|                          |
  |                       |                          |
  | 407 Proxy Auth        |                          |
  |<----------------------|                          |
  |                       |                          |
  | Request + Auth Header |                          |
  |---------------------->|                          |
  |                       |---(if auth valid)------->|
  |                       |                          |
  ```

## Step2: Add TLS Support for HTTPS Interception
- [ ] Generate a self-signed CA certificate
- [ ] Sign leaf certificates dynamically for different domains
- [ ] Use `crypto/tls` and `mitmproxy` techniques to decrypt traffic
- [ ] Parse Client Hello message
- [ ] Deal with 0-RTT

## Step3: Parse and Modify Requests and Responses
- [ ] Decode HTTP headers, body, and parameters
- [ ] Allow modifications (e.g., inject headers, replace payloads)
- [ ] Support traffic inspection for debugging

## Step4: Implement Forwarding Logic
- [x] Set up a TCP tunneling mechanism for HTTPS
- [x] May use golang built-in support `io.Copy`

## Step5: Logging and Debugging
- [ ] Store HTTP request/response details in database
- [ ] Implement real-time monitoring via WebSocket or API

## Step6: Support Custom Rules and Filters
- [ ] Allow users to define rules (e.g. blocking by domains, modify content)

## Step7: Improve Performance and Security
- [ ] Use goroutines go handle concurrent requests
- [ ] Optimize for **high throughput** with connection pooling



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
