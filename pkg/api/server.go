package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/clau/email_verifier/pkg/config"
	"github.com/clau/email_verifier/pkg/verifier"
	"github.com/gorilla/mux"
)

// Server represents the API server
type Server struct {
	router   *mux.Router
	verifier *verifier.Verifier
	config   *config.Config
}

// VerifyRequest represents a request to verify an email
type VerifyRequest struct {
	Email string `json:"email"`
}

// VerifyResponse represents the response from verifying an email
type VerifyResponse struct {
	Email              string `json:"email"`
	VerificationStatus string `json:"verification_status"`
	ConfidenceScore    int    `json:"confidence_score"`
	ProcessedAt        string `json:"processed_at"`
}

// BatchVerifyRequest represents a request to verify multiple emails
type BatchVerifyRequest struct {
	Emails []string `json:"emails"`
}

// BatchVerifyResponse represents the response from verifying multiple emails
type BatchVerifyResponse struct {
	Results []VerifyResponse `json:"results"`
}

// NewServer creates a new API server
func NewServer(cfg *config.Config) *Server {
	v := verifier.New(cfg)
	r := mux.NewRouter()

	server := &Server{
		router:   r,
		verifier: v,
		config:   cfg,
	}

	// Register routes
	r.HandleFunc("/health", server.healthHandler).Methods("GET")
	r.HandleFunc("/verify", server.verifyHandler).Methods("POST")
	r.HandleFunc("/batch-verify", server.batchVerifyHandler).Methods("POST")
	r.HandleFunc("/google-sheets", server.handleGoogleSheetsRequest).Methods("POST", "OPTIONS")

	// Add middleware for logging and CORS
	r.Use(loggingMiddleware)
	r.Use(corsMiddleware)

	return server
}

// Start starts the API server
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting API server on %s", addr)
	return http.ListenAndServe(addr, s.router)
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// verifyHandler handles email verification requests
func (s *Server) verifyHandler(w http.ResponseWriter, r *http.Request) {
	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(req.Email)
	result, err := s.verifier.VerifyWithRetry(email)
	if err != nil {
		log.Printf("Error verifying email %s: %v", email, err)
		http.Error(w, "Error verifying email", http.StatusInternalServerError)
		return
	}

	status, score := s.verifier.DetermineStatus(result, email)

	response := VerifyResponse{
		Email:              email,
		VerificationStatus: status,
		ConfidenceScore:    score,
		ProcessedAt:        time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// batchVerifyHandler handles batch email verification requests
func (s *Server) batchVerifyHandler(w http.ResponseWriter, r *http.Request) {
	var req BatchVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Emails) == 0 {
		http.Error(w, "Emails list is required", http.StatusBadRequest)
		return
	}

	if len(req.Emails) > 100 {
		http.Error(w, "Maximum 100 emails allowed per batch", http.StatusBadRequest)
		return
	}

	results := make([]VerifyResponse, 0, len(req.Emails))

	for _, email := range req.Emails {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}

		result, err := s.verifier.VerifyWithRetry(email)
		if err != nil {
			log.Printf("Error verifying email %s: %v", email, err)
			results = append(results, VerifyResponse{
				Email:              email,
				VerificationStatus: "error",
				ConfidenceScore:    0,
				ProcessedAt:        time.Now().Format(time.RFC3339),
			})
			continue
		}

		status, score := s.verifier.DetermineStatus(result, email)

		results = append(results, VerifyResponse{
			Email:              email,
			VerificationStatus: status,
			ConfidenceScore:    score,
			ProcessedAt:        time.Now().Format(time.RFC3339),
		})
	}

	response := BatchVerifyResponse{
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// loggingMiddleware logs all requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
