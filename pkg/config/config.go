package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the configurable parameters
type Config struct {
	InputFile         string         `yaml:"input_file"`
	InputType         string         `yaml:"input_type"`
	OutputFile        string         `yaml:"output_file"`
	OutputType        string         `yaml:"output_type"`
	ValidThreshold    int            `yaml:"valid_threshold"`
	RiskyThreshold    int            `yaml:"risky_threshold"`
	DefaultRiskyScore int            `yaml:"default_risky_score"`
	MaxRetries        int            `yaml:"max_retries"`
	InitialBackoff    time.Duration  `yaml:"initial_backoff"`
	NumWorkers        int            `yaml:"num_workers"`
	ScoringWeights    ScoringWeights `yaml:"scoring_weights"`
}

// ScoringWeights to manage individual weights in config
type ScoringWeights struct {
	HasMxRecords     int `yaml:"has_mx_records"`
	ReachableYes     int `yaml:"reachable_yes"`
	ReachableUnknown int `yaml:"reachable_unknown"`
	RoleAccount      int `yaml:"role_account"`
	FreeProvider     int `yaml:"free_provider"`
	Suggestion       int `yaml:"suggestion"`
}

// LoadConfig loads and validates configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(configFile, config)
	if err != nil {
		return nil, err
	}

	// Validate input and output types
	config.InputType = strings.ToLower(config.InputType)
	config.OutputType = strings.ToLower(config.OutputType)

	if config.InputType != "csv" && config.InputType != "xlsx" {
		return nil, fmt.Errorf("invalid input_type in config.yaml: %s. Must be 'csv' or 'xlsx'", config.InputType)
	}
	if config.OutputType != "csv" && config.OutputType != "xlsx" {
		return nil, fmt.Errorf("invalid output_type in config.yaml: %s. Must be 'csv' or 'xlsx'", config.OutputType)
	}

	return config, nil
}
