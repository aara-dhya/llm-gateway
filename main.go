package main

import (
	"context" // Add this
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings" // Add this
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

// LLMResponse matches the standard OpenAI JSON response structure
type LLMResponse struct {
	Model string `json:"model"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// RequestLog represents a metric record returned by the database
type RequestLog struct {
	ID               int       `json:"id"`
	Timestamp        time.Time `json:"timestamp"`
	Model            string    `json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
}

func initDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		// Default to port 5433 if you mapped Docker host port 5433, or 5432
		connStr = "host=localhost port=5433 user=postgres password=postgres dbname=llm_gateway sslmode=disable"
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatalf("PostgreSQL connection check failed: %v", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS requests (
		id SERIAL PRIMARY KEY,
		timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		model VARCHAR(255),
		prompt_tokens INT,
		completion_tokens INT,
		total_tokens INT
	);`

	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}
	log.Println("PostgreSQL database initialized successfully.")
}

func logMetrics(model string, promptTokens, completionTokens, totalTokens int) {
	query := `INSERT INTO requests (model, prompt_tokens, completion_tokens, total_tokens) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, model, promptTokens, completionTokens, totalTokens)
	if err != nil {
		log.Printf("Failed to log metrics to PostgreSQL: %v", err)
	} else {
		log.Printf("Logged request -> Model: %s | Tokens: %d (Prompt: %d, Completion: %d)",
			model, totalTokens, promptTokens, completionTokens)
	}
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received proxy request: %s %s", r.Method, r.URL.Path)

	targetURL := "https://api.openai.com/v1/chat/completions"

	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to reach LLM server", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read upstream response", http.StatusInternalServerError)
		return
	}

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(bodyBytes)

	if resp.StatusCode == http.StatusOK {
		var llmResp LLMResponse
		if err := json.Unmarshal(bodyBytes, &llmResp); err == nil {
			logMetrics(
				llmResp.Model,
				llmResp.Usage.PromptTokens,
				llmResp.Usage.CompletionTokens,
				llmResp.Usage.TotalTokens,
			)
		} else {
			log.Printf("Failed to parse response JSON: %v", err)
		}
	}
}

// handleMetrics queries PostgreSQL for request logs and serves them as JSON
func handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Fetch up to 100 most recent requests from PostgreSQL
	query := `SELECT id, timestamp, model, prompt_tokens, completion_tokens, total_tokens 
	          FROM requests 
	          ORDER BY timestamp DESC 
	          LIMIT 100`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying metrics: %v", err)
		http.Error(w, "Failed to fetch metrics", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	logs := []RequestLog{}
	for rows.Next() {
		var l RequestLog
		if err := rows.Scan(&l.ID, &l.Timestamp, &l.Model, &l.PromptTokens, &l.CompletionTokens, &l.TotalTokens); err != nil {
			log.Printf("Error scanning row: %v", err)
			http.Error(w, "Failed to read metric record", http.StatusInternalServerError)
			return
		}
		logs = append(logs, l)
	}

	// Set header so callers know this is JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract API Key from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Missing or invalid token", http.StatusUnauthorized)
			return
		}
		apiKey := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Database Lookup & Balance Check
		var userID int
		var tokenBalance int
		err := db.QueryRow("SELECT id, token_balance FROM users WHERE api_key = $1", apiKey).Scan(&userID, &tokenBalance)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid API key", http.StatusUnauthorized)
			return
		}

		if tokenBalance <= 0 {
			http.Error(w, "Forbidden: Insufficient token balance", http.StatusForbidden)
			return
		}

		// 3. Inject the user_id into the request context for logging later
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	initDB()
	defer db.Close()

	// OpenAI Reverse Proxy Route
	http.Handle("/v1/chat/completions", authMiddleware(http.HandlerFunc(handleProxy)))
	// Observability Metrics REST Route
	http.HandleFunc("/api/metrics", handleMetrics)

	port := ":8080"
	log.Printf("Starting LLM API Gateway on port %s...", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
