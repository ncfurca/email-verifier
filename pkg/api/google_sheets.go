package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// GoogleSheetsRequest represents a request from Google Sheets
type GoogleSheetsRequest struct {
	Emails []string `json:"emails"`
}

// GoogleSheetsResponse represents a response for Google Sheets
type GoogleSheetsResponse struct {
	Results []GoogleSheetsResult `json:"results"`
}

// GoogleSheetsResult represents a single email verification result for Google Sheets
type GoogleSheetsResult struct {
	Email              string `json:"email"`
	VerificationStatus string `json:"verification_status"`
	ConfidenceScore    int    `json:"confidence_score"`
	ProcessedAt        string `json:"processed_at"`
}

// handleGoogleSheetsRequest handles requests from Google Sheets
func (s *Server) handleGoogleSheetsRequest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for pre-flight requests
	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Set CORS headers for the main request
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Parse the request
	var req GoogleSheetsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate the request
	if len(req.Emails) == 0 {
		http.Error(w, "No emails provided", http.StatusBadRequest)
		return
	}

	// Limit the number of emails to process
	if len(req.Emails) > 100 {
		http.Error(w, "Too many emails (maximum 100)", http.StatusBadRequest)
		return
	}

	// Process the emails
	results := make([]GoogleSheetsResult, 0, len(req.Emails))
	for _, email := range req.Emails {
		email = strings.TrimSpace(email)
		if email == "" {
			continue
		}

		// Verify the email
		result, err := s.verifier.VerifyWithRetry(email)
		if err != nil {
			log.Printf("Error verifying email %s: %v", email, err)
			results = append(results, GoogleSheetsResult{
				Email:              email,
				VerificationStatus: "error",
				ConfidenceScore:    0,
				ProcessedAt:        time.Now().Format(time.RFC3339),
			})
			continue
		}

		// Determine the status
		status, score := s.verifier.DetermineStatus(result, email)

		// Add the result
		results = append(results, GoogleSheetsResult{
			Email:              email,
			VerificationStatus: status,
			ConfidenceScore:    score,
			ProcessedAt:        time.Now().Format(time.RFC3339),
		})
	}

	// Return the results
	response := GoogleSheetsResponse{
		Results: results,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
