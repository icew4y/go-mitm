package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

func handleRequestAndRedirect(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{
		Timeout: time.Duration(time.Second * 30),
	}

	// Just simply start a new request with the original args
	newReq, err := http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		log.Printf("http.NewRequest: %v", err)
		http.Error(w, "http.NewRequest", http.StatusInternalServerError)
		return
	}

	newReq.Header = r.Header

	// Send it

	resp, err := client.Do(newReq)
	if err != nil {
		log.Printf("client.Do: %v", err)
		http.Error(w, "client.Do", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	// Copy the headers

	for key, val := range resp.Header {
		for _, v := range val {
			w.Header().Add(key, v)
		}
	}

	w.WriteHeader(resp.StatusCode)

	// Copy response body

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("io.Copy: %v", err)
		http.Error(w, "io.Copy", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", handleRequestAndRedirect)
	log.Println("Starting HTTP Proxy on :8080")

	// Listen on port 8080
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Failed to start the HTTP server on :8080: %v", err)
	}
}
