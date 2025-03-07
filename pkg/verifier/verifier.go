package verifier

import (
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"

	emailverifier "github.com/AfterShip/email-verifier"
	"github.com/clau/email_verifier/pkg/config"
	"github.com/clau/email_verifier/pkg/utils"
)

// Result represents the result of email verification
type Result struct {
	Email              string `json:"email"`
	VerificationStatus string `json:"verification_status"` // valid, invalid, risky
	ConfidenceScore    int    `json:"confidence_score"`    // 0 to 100
}

// Verifier handles email verification operations
type Verifier struct {
	config      *config.Config
	pool        *sync.Pool
	rateLimiter *rateLimiter
	mu          sync.Mutex // Mutex for thread-safe operations
}

type rateLimiter struct {
	mu       sync.Mutex
	lastCall time.Time
	interval time.Duration
}

func newRateLimiter(interval time.Duration) *rateLimiter {
	return &rateLimiter{
		interval: interval,
	}
}

func (rl *rateLimiter) wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if elapsed := now.Sub(rl.lastCall); elapsed < rl.interval {
		time.Sleep(rl.interval - elapsed)
	}
	rl.lastCall = time.Now()
}

// New creates a new email verifier instance
func New(cfg *config.Config) *Verifier {
	return &Verifier{
		config: cfg,
		pool: &sync.Pool{
			New: func() interface{} {
				return emailverifier.NewVerifier().
					EnableSMTPCheck().
					EnableDomainSuggest()
			},
		},
		rateLimiter: newRateLimiter(100 * time.Millisecond), // 10 requests per second
	}
}

// VerifyWithRetry attempts to verify an email with retries
func (v *Verifier) VerifyWithRetry(email string) (*emailverifier.Result, error) {
	var result *emailverifier.Result
	var err error

	// Get a verifier from the pool
	verifier := v.pool.Get().(*emailverifier.Verifier)
	defer v.pool.Put(verifier)

	for attempt := 0; attempt <= v.config.MaxRetries; attempt++ {
		// Apply rate limiting
		v.rateLimiter.wait()

		// Perform verification with proper error handling
		func() {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic during verification: %v", r)
				}
			}()
			result, err = verifier.Verify(email)
		}()

		if err == nil {
			return result, nil
		}

		if isRetryableError(err) {
			backoffDuration := v.config.InitialBackoff * time.Duration(math.Pow(2, float64(attempt)))
			log.Printf("Attempt %d: Error verifying email %s: %v. Retrying in %v...", attempt+1, email, err, backoffDuration)
			time.Sleep(backoffDuration)
		} else {
			return nil, err
		}
	}

	return nil, err
}

// DetermineStatus calculates the verification status and confidence score
func (v *Verifier) DetermineStatus(result *emailverifier.Result, email string) (string, int) {
	if result == nil {
		log.Printf("Warning: Nil result for email %s. Marking as invalid.", email)
		return "invalid", 0
	}

	fmt.Printf("\n--- Verification Details for %s ---\n", email)
	fmt.Printf("Syntax Valid: %t\n", result.Syntax.Valid)
	fmt.Printf("Disposable: %t\n", result.Disposable)
	fmt.Printf("Has MX Records: %t\n", result.HasMxRecords)
	fmt.Printf("Reachable: %s\n", result.Reachable)
	fmt.Printf("Role Account: %t\n", result.RoleAccount)
	fmt.Printf("Free Provider: %t\n", result.Free)
	fmt.Printf("Suggestion: %s\n", result.Suggestion)

	confidenceScore := 50 // Base confidence

	// Rigid Invalid Conditions (highest priority)
	if !result.Syntax.Valid {
		fmt.Println("Status: invalid - Invalid Syntax")
		return "invalid", 0
	}
	if result.Reachable == "no" {
		fmt.Println("Status: invalid - Reachable: no")
		return "invalid", 0
	}
	if result.Disposable {
		fmt.Println("Status: invalid - Disposable Email")
		return "invalid", 10
	}

	// Positive signals
	if result.HasMxRecords {
		confidenceScore += v.config.ScoringWeights.HasMxRecords
	}
	if result.Reachable == "yes" {
		confidenceScore += v.config.ScoringWeights.ReachableYes
	}

	// Negative signals
	if result.Reachable == "unknown" {
		confidenceScore += v.config.ScoringWeights.ReachableUnknown
	}
	if result.RoleAccount {
		confidenceScore += v.config.ScoringWeights.RoleAccount
	}
	if result.Free {
		confidenceScore += v.config.ScoringWeights.FreeProvider
	}
	if result.Suggestion != "" {
		confidenceScore += v.config.ScoringWeights.Suggestion
	}

	confidenceScore = utils.Max(0, utils.Min(confidenceScore, 100))

	var verificationStatus string
	switch {
	case confidenceScore >= v.config.ValidThreshold:
		verificationStatus = "valid"
	case confidenceScore >= v.config.RiskyThreshold:
		verificationStatus = "risky"
	default:
		verificationStatus = "invalid"
	}

	return verificationStatus, confidenceScore
}

// isRetryableError checks if an error is considered retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "i/o error") ||
		strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "network is unreachable")
}
