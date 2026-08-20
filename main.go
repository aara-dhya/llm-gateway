package main

import (
	"fmt"
	"log"
	"net/http"
)

// handleProxy is our placeholder handler.
// Later, this will forward requests to the LLM.
func handleProxy(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request method and path for now
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)

	// Send a simple text response to the client
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Gateway is running. Ready to proxy requests.")
}

func main() {
	// Register our handler function to intercept all requests to /v1/chat/completions
	// (This is the standard OpenAI endpoint format)
	http.HandleFunc("/v1/chat/completions", handleProxy)

	port := ":8080"
	log.Printf("Starting LLM API Gateway on port %s...", port)

	// Start the server. ListenAndServe blocks forever unless an error occurs.
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
