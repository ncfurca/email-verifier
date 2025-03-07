package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/clau/email_verifier/pkg/config"
	"github.com/clau/email_verifier/pkg/io"
	"github.com/clau/email_verifier/pkg/verifier"
)

type progress struct {
	mu         sync.Mutex
	total      int
	processed  int
	valid      int
	risky      int
	invalid    int
	errors     int
	startTime  time.Time
	lastUpdate time.Time
}

func (p *progress) update(status string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.processed++
	switch status {
	case "valid":
		p.valid++
	case "risky":
		p.risky++
	case "invalid":
		p.invalid++
	case "error":
		p.errors++
	}

	// Update progress every second
	if time.Since(p.lastUpdate) >= time.Second {
		elapsed := time.Since(p.startTime)
		rate := float64(p.processed) / elapsed.Seconds()
		fmt.Printf("\rProgress: %d/%d (%.1f%%), Rate: %.1f/s, Valid: %d, Risky: %d, Invalid: %d, Errors: %d",
			p.processed, p.total,
			float64(p.processed)*100/float64(p.total),
			rate,
			p.valid, p.risky, p.invalid, p.errors)
		p.lastUpdate = time.Now()
	}
}

// findFile attempts to find a file with case-insensitive matching
func findFile(filename string) (string, error) {
	// First, try the exact filename
	if _, err := os.Stat(filename); err == nil {
		return filename, nil
	}

	// If not found, try case-insensitive search in the directory
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)

	// If the directory is empty or ".", use the current directory
	if dir == "" || dir == "." {
		dir = "."
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("error reading directory %s: %v", dir, err)
	}

	for _, entry := range entries {
		if strings.EqualFold(entry.Name(), base) {
			return filepath.Join(dir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("file not found: %s", filename)
}

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Find input file with case-insensitive matching
	inputFile, err := findFile(cfg.InputFile)
	if err != nil {
		log.Fatalf("Error finding input file: %v", err)
	}

	// Update config with the actual file path
	cfg.InputFile = inputFile

	// Read input records
	records, err := io.ReadRecords(cfg.InputFile, cfg.InputType)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	if len(records) == 0 {
		log.Fatalf("No records found in input file: %s", cfg.InputFile)
	}

	// Validate that records have email field (case-insensitive)
	for i, record := range records {
		hasEmail := false
		for key := range record {
			if strings.EqualFold(key, "email") {
				hasEmail = true
				break
			}
		}
		if !hasEmail {
			log.Fatalf("Record at index %d is missing the required 'email' field", i)
		}
	}

	// Initialize progress tracking
	progress := &progress{
		total:      len(records),
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}

	// Initialize verifier
	v := verifier.New(cfg)

	// Process records concurrently
	var wg sync.WaitGroup
	recordsChan := make(chan map[string]string)
	resultsChan := make(chan verifier.Result)

	// Start workers
	for i := 0; i < cfg.NumWorkers; i++ {
		wg.Add(1)
		go worker(recordsChan, resultsChan, &wg, v, progress)
	}

	// Send records to workers
	go func() {
		for _, record := range records {
			recordsChan <- record
		}
		close(recordsChan)
	}()

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	resultsMap := make(map[string]verifier.Result)
	for result := range resultsChan {
		// Use lowercase email as key for case-insensitive matching
		resultsMap[strings.ToLower(strings.TrimSpace(result.Email))] = result
	}

	// Prepare results in original order
	results := make([]verifier.Result, 0, len(records))
	for _, record := range records {
		// Get email field (case-insensitive)
		var email string
		for key, value := range record {
			if strings.EqualFold(key, "email") {
				email = strings.TrimSpace(value)
				break
			}
		}

		if email == "" {
			continue
		}

		// Look up result by lowercase email for case-insensitive matching
		if result, ok := resultsMap[strings.ToLower(email)]; ok {
			results = append(results, result)
		} else {
			log.Printf("Warning: No verification result found for email: %s. Setting to invalid.", email)
			results = append(results, verifier.Result{
				Email:              email,
				VerificationStatus: "invalid",
				ConfidenceScore:    0,
			})
		}
	}

	// Write results
	err = io.WriteResults(cfg.OutputFile, cfg.OutputType, records, results)
	if err != nil {
		log.Fatalf("Error writing results: %v", err)
	}

	// Print final statistics
	elapsed := time.Since(progress.startTime)
	fmt.Printf("\n\nVerification completed in %v\n", elapsed)
	fmt.Printf("Total processed: %d\n", progress.processed)
	fmt.Printf("Valid: %d (%.1f%%)\n", progress.valid, float64(progress.valid)*100/float64(progress.processed))
	fmt.Printf("Risky: %d (%.1f%%)\n", progress.risky, float64(progress.risky)*100/float64(progress.processed))
	fmt.Printf("Invalid: %d (%.1f%%)\n", progress.invalid, float64(progress.invalid)*100/float64(progress.processed))
	fmt.Printf("Errors: %d (%.1f%%)\n", progress.errors, float64(progress.errors)*100/float64(progress.processed))
	fmt.Printf("Results saved to %s\n", cfg.OutputFile)
}

func worker(recordsChan <-chan map[string]string, resultsChan chan<- verifier.Result, wg *sync.WaitGroup, v *verifier.Verifier, progress *progress) {
	defer wg.Done()

	for record := range recordsChan {
		// Get email field (case-insensitive)
		var email string
		for key, value := range record {
			if strings.EqualFold(key, "email") {
				email = strings.TrimSpace(value)
				break
			}
		}

		if email == "" {
			continue
		}

		result, err := v.VerifyWithRetry(email)
		if err != nil {
			log.Printf("Error verifying email %s after retries: %v. Marking as invalid.", email, err)
			resultsChan <- verifier.Result{
				Email:              email,
				VerificationStatus: "invalid",
				ConfidenceScore:    0,
			}
			progress.update("error")
			continue
		}

		verificationStatus, confidenceScore := v.DetermineStatus(result, email)
		resultsChan <- verifier.Result{
			Email:              email,
			VerificationStatus: verificationStatus,
			ConfidenceScore:    confidenceScore,
		}
		progress.update(verificationStatus)
	}
}
