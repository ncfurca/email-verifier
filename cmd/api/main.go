package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/clau/email_verifier/pkg/api"
	"github.com/clau/email_verifier/pkg/config"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	port := flag.Int("port", 8080, "Port to run the API server on")
	flag.Parse()

	// Find config file with case-insensitive matching
	configPath, err := findFile(*configFile)
	if err != nil {
		log.Printf("Warning: Could not find config file at %s: %v", *configFile, err)
		log.Printf("Using default configuration")

		// Create a default config
		cfg := createDefaultConfig()

		// Start the API server
		server := api.NewServer(cfg)
		log.Fatal(server.Start(*port))
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Start the API server
	server := api.NewServer(cfg)
	log.Fatal(server.Start(*port))
}

// createDefaultConfig creates a default configuration for the API server
func createDefaultConfig() *config.Config {
	return &config.Config{
		InputFile:         "input.csv",
		OutputFile:        "output.csv",
		InputType:         "csv",
		OutputType:        "csv",
		ValidThreshold:    80,
		RiskyThreshold:    60,
		DefaultRiskyScore: 50,
		MaxRetries:        3,
		InitialBackoff:    time.Second,
		NumWorkers:        10,
		ScoringWeights: config.ScoringWeights{
			HasMxRecords:     20,
			ReachableYes:     40,
			ReachableUnknown: 20,
			RoleAccount:      -10,
			FreeProvider:     -5,
			Suggestion:       -10,
		},
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
		if entry.Name() == base || strings.EqualFold(entry.Name(), base) {
			return filepath.Join(dir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("file not found: %s", filename)
}
