package main

import (
	"io"
	"log"
	"net/http"
)

func handleProxy(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)

	// 1. Create a new HTTP request pointing to OpenAI's API
	// We use the same method (POST) and pass along the exact body from the incoming request.
	proxyReq, err := http.NewRequest(r.Method, "https://api.openai.com/v1/chat/completions", r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	// 2. Copy all headers from the incoming request to our proxy request.
	// This is critical because it passes along the "Authorization: Bearer <token>"
	// and "Content-Type: application/json" headers.
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// 3. Execute the request using Go's default HTTP client
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to reach OpenAI", http.StatusBadGateway)
		return
	}
	// Defer closing the response body so memory is freed when the function exits
	defer resp.Body.Close()

	// 4. Copy the response headers from OpenAI back to our client (e.g., Content-Type)
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// 5. Pass along the exact HTTP status code that OpenAI returned (e.g., 200 OK, 401 Unauthorized)
	w.WriteHeader(resp.StatusCode)

	// 6. Stream the response body from OpenAI directly to our client
	io.Copy(w, resp.Body)
}

func main() {
	http.HandleFunc("/v1/chat/completions", handleProxy)

	port := ":8080"
	log.Printf("Starting LLM API Gateway on port %s...", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
