package io

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/clau/email_verifier/pkg/verifier"
	"github.com/tealeg/xlsx"
)

// WriteResults writes verification results to either CSV or XLSX file
func WriteResults(filePath, fileType string, records []map[string]string, results []verifier.Result) error {
	switch strings.ToLower(fileType) {
	case "csv":
		return writeResultsToCSV(filePath, records, results)
	case "xlsx":
		return writeResultsToXLSX(filePath, records, results)
	default:
		return fmt.Errorf("unsupported file type: %s", fileType)
	}
}

// getHeaders extracts all unique headers from all records
func getHeaders(records []map[string]string) []string {
	// Use a map to track all unique keys across all records
	headerMap := make(map[string]bool)

	// Collect all unique keys from all records (using lowercase for consistency)
	for _, record := range records {
		for key := range record {
			// Store lowercase version of the key
			headerMap[strings.ToLower(key)] = true
		}
	}

	// Convert map keys to a slice
	headers := make([]string, 0, len(headerMap))
	for key := range headerMap {
		headers = append(headers, key)
	}

	// Sort headers alphabetically for consistency, but ensure 'email' is first if present
	sort.Slice(headers, func(i, j int) bool {
		if headers[i] == "email" {
			return true
		}
		if headers[j] == "email" {
			return false
		}
		return headers[i] < headers[j]
	})

	return headers
}

// getOutputHeaders returns all original headers plus verification result headers
func getOutputHeaders(originalHeaders []string) []string {
	// Create a copy of the original headers
	headers := make([]string, len(originalHeaders))
	copy(headers, originalHeaders)

	// Add verification result headers (lowercase)
	verificationHeaders := []string{
		"verification status",
		"confidence score",
	}

	// Check if these headers already exist in the original data
	for _, vh := range verificationHeaders {
		exists := false
		for _, h := range headers {
			if strings.ToLower(h) == vh {
				exists = true
				break
			}
		}

		if !exists {
			headers = append(headers, vh)
		} else {
			// If header already exists, use a modified name to avoid collision
			headers = append(headers, vh+" (verification)")
		}
	}

	return headers
}

// getLowercaseValue gets a value from a record using case-insensitive key matching
func getLowercaseValue(record map[string]string, key string) string {
	// First try direct lookup with the lowercase key
	if value, exists := record[key]; exists {
		return value
	}

	// If not found, try case-insensitive lookup
	lowercaseKey := strings.ToLower(key)
	for k, v := range record {
		if strings.ToLower(k) == lowercaseKey {
			return v
		}
	}

	return "" // Return empty string if not found
}

// writeResultsToCSV writes the verification results to a CSV file
func writeResultsToCSV(filePath string, records []map[string]string, results []verifier.Result) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Get all unique headers from records
	originalHeaders := getHeaders(records)

	// Get output headers including verification results
	outputHeaders := getOutputHeaders(originalHeaders)

	// Write header row (all lowercase)
	lowercaseHeaders := make([]string, len(outputHeaders))
	for i, header := range outputHeaders {
		lowercaseHeaders[i] = strings.ToLower(header)
	}

	if err := writer.Write(lowercaseHeaders); err != nil {
		return err
	}

	// Write data rows
	for i, record := range records {
		// Create a slice for the row with all fields
		row := make([]string, len(outputHeaders))

		// Add original record fields
		for j, header := range originalHeaders {
			row[j] = getLowercaseValue(record, header)
		}

		// Add verification results
		var result verifier.Result
		email := getLowercaseValue(record, "email")
		email = strings.TrimSpace(email)

		if i < len(results) && strings.EqualFold(email, results[i].Email) {
			result = results[i]
		} else {
			// Try to find matching result by email (case-insensitive)
			found := false
			for _, r := range results {
				if strings.EqualFold(email, r.Email) {
					result = r
					found = true
					break
				}
			}

			if !found {
				log.Printf("Warning: No matching result found for email: %s. Setting to invalid.", email)
				result = verifier.Result{Email: email, VerificationStatus: "invalid", ConfidenceScore: 0}
			}
		}

		// Add verification status and confidence score at the end
		verificationStartIdx := len(originalHeaders)
		row[verificationStartIdx] = result.VerificationStatus
		row[verificationStartIdx+1] = fmt.Sprintf("%d", result.ConfidenceScore)

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// writeResultsToXLSX writes the verification results to an Excel file
func writeResultsToXLSX(filePath string, records []map[string]string, results []verifier.Result) error {
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("verified leads")
	if err != nil {
		return err
	}

	// Get all unique headers from records
	originalHeaders := getHeaders(records)

	// Get output headers including verification results
	outputHeaders := getOutputHeaders(originalHeaders)

	// Write header row (all lowercase)
	headerRow := sheet.AddRow()
	for _, header := range outputHeaders {
		headerCell := headerRow.AddCell()
		headerCell.SetString(strings.ToLower(header))
	}

	// Write data rows
	for i, record := range records {
		row := sheet.AddRow()

		// Add original record fields
		for _, header := range originalHeaders {
			cell := row.AddCell()
			cell.SetString(getLowercaseValue(record, header))
		}

		// Add verification results
		var result verifier.Result
		email := getLowercaseValue(record, "email")
		email = strings.TrimSpace(email)

		if i < len(results) && strings.EqualFold(email, results[i].Email) {
			result = results[i]
		} else {
			// Try to find matching result by email (case-insensitive)
			found := false
			for _, r := range results {
				if strings.EqualFold(email, r.Email) {
					result = r
					found = true
					break
				}
			}

			if !found {
				log.Printf("Warning: No matching result found for email: %s. Setting to invalid.", email)
				result = verifier.Result{Email: email, VerificationStatus: "invalid", ConfidenceScore: 0}
			}
		}

		// Add verification status and confidence score
		statusCell := row.AddCell()
		statusCell.SetString(result.VerificationStatus)
		scoreCell := row.AddCell()
		scoreCell.SetInt(result.ConfidenceScore)
	}

	return file.Save(filePath)
}
