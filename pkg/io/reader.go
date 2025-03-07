package io

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/tealeg/xlsx"
)

// ReadRecords reads records from either CSV or XLSX file based on file type
func ReadRecords(filePath, fileType string) ([]map[string]string, error) {
	switch strings.ToLower(fileType) {
	case "csv":
		return readRecordsFromCSV(filePath)
	case "xlsx":
		return readRecordsFromXLSX(filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
}

// normalizeHeaders converts all headers to lowercase for case-insensitive matching
func normalizeHeaders(headers []string) []string {
	normalized := make([]string, len(headers))
	for i, header := range headers {
		normalized[i] = strings.ToLower(strings.TrimSpace(header))
	}
	return normalized
}

// readRecordsFromCSV reads records from a CSV file
func readRecordsFromCSV(filePath string) ([]map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, fmt.Errorf("empty CSV file: %s", filePath)
	}

	// Get original headers for preserving case in output
	originalHeaders := rows[0]

	// Normalize headers for case-insensitive matching
	normalizedHeaders := normalizeHeaders(originalHeaders)

	records := make([]map[string]string, 0, len(rows)-1)
	for _, row := range rows[1:] {
		record := make(map[string]string)
		for i, value := range row {
			if i < len(normalizedHeaders) {
				// Store with lowercase key for consistent access
				record[normalizedHeaders[i]] = value

				// Also store with original case for backward compatibility
				if originalHeaders[i] != normalizedHeaders[i] {
					record[originalHeaders[i]] = value
				}
			}
		}
		records = append(records, record)
	}
	return records, nil
}

// readRecordsFromXLSX reads records from an XLSX file
func readRecordsFromXLSX(filePath string) ([]map[string]string, error) {
	xlFile, err := xlsx.OpenFile(filePath)
	if err != nil {
		return nil, err
	}

	if len(xlFile.Sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in XLSX file: %s", filePath)
	}

	sheet := xlFile.Sheets[0] // Assuming data is in the first sheet
	if sheet == nil {
		return nil, fmt.Errorf("sheet not found in XLSX file")
	}

	if len(sheet.Rows) == 0 {
		return nil, fmt.Errorf("empty sheet in XLSX file: %s", filePath)
	}

	// Get original headers for preserving case in output
	originalHeaders := make([]string, 0)
	if len(sheet.Rows) > 0 && len(sheet.Rows[0].Cells) > 0 {
		for _, cell := range sheet.Rows[0].Cells {
			originalHeaders = append(originalHeaders, cell.Value)
		}
	}

	// Normalize headers for case-insensitive matching
	normalizedHeaders := normalizeHeaders(originalHeaders)

	records := make([]map[string]string, 0)
	for rowIndex, row := range sheet.Rows {
		if rowIndex == 0 { // Skip header row
			continue
		}

		record := make(map[string]string)
		for cellIndex, cell := range row.Cells {
			if cellIndex < len(normalizedHeaders) {
				value := cell.Value

				// Store with lowercase key for consistent access
				record[normalizedHeaders[cellIndex]] = value

				// Also store with original case for backward compatibility
				if originalHeaders[cellIndex] != normalizedHeaders[cellIndex] {
					record[originalHeaders[cellIndex]] = value
				}
			}
		}
		records = append(records, record)
	}
	return records, nil
}
