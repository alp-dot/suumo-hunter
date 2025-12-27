package models

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// CSV header columns in order.
var csvHeaders = []string{
	"id",
	"name",
	"address",
	"age",
	"floor",
	"rent",
	"management_fee",
	"deposit",
	"key_money",
	"layout",
	"area",
	"walk_minutes",
	"url",
}

// LoadFromCSV reads properties from a CSV file.
// The CSV must have a header row matching the expected columns.
func LoadFromCSV(r io.Reader) ([]Property, error) {
	reader := csv.NewReader(r)

	// Read header
	header, err := reader.Read()
	if err != nil {
		if err == io.EOF {
			return []Property{}, nil
		}
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Create column index map
	colIndex := make(map[string]int)
	for i, col := range header {
		colIndex[col] = i
	}

	// Verify required columns exist
	for _, required := range csvHeaders {
		if _, ok := colIndex[required]; !ok {
			return nil, fmt.Errorf("missing required column: %s", required)
		}
	}

	var properties []Property
	lineNum := 1 // Header is line 1

	for {
		lineNum++
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV line %d: %w", lineNum, err)
		}

		prop := recordToProperty(record, colIndex)
		properties = append(properties, prop)
	}

	return properties, nil
}

// recordToProperty converts a CSV record to a Property struct.
func recordToProperty(record []string, colIndex map[string]int) Property {
	getField := func(name string) string {
		if idx, ok := colIndex[name]; ok && idx < len(record) {
			return record[idx]
		}
		return ""
	}

	age, _ := strconv.Atoi(getField("age"))
	floor, _ := strconv.Atoi(getField("floor"))
	rent, _ := strconv.ParseFloat(getField("rent"), 64)
	managementFee, _ := strconv.ParseFloat(getField("management_fee"), 64)
	area, _ := strconv.ParseFloat(getField("area"), 64)
	walkMinutes, _ := strconv.Atoi(getField("walk_minutes"))

	return Property{
		ID:            getField("id"),
		Name:          getField("name"),
		Address:       getField("address"),
		Age:           age,
		Floor:         floor,
		Rent:          rent,
		ManagementFee: managementFee,
		Deposit:       getField("deposit"),
		KeyMoney:      getField("key_money"),
		Layout:        getField("layout"),
		Area:          area,
		WalkMinutes:   walkMinutes,
		URL:           getField("url"),
	}
}

// SaveToCSV writes properties to a CSV file.
// The CSV will have a header row followed by data rows.
func SaveToCSV(w io.Writer, properties []Property) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	if err := writer.Write(csvHeaders); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for i, prop := range properties {
		record := propertyToRecord(prop)
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV row %d: %w", i+1, err)
		}
	}

	return nil
}

// propertyToRecord converts a Property struct to a CSV record.
func propertyToRecord(p Property) []string {
	return []string{
		p.ID,
		p.Name,
		p.Address,
		strconv.Itoa(p.Age),
		strconv.Itoa(p.Floor),
		strconv.FormatFloat(p.Rent, 'f', -1, 64),
		strconv.FormatFloat(p.ManagementFee, 'f', -1, 64),
		p.Deposit,
		p.KeyMoney,
		p.Layout,
		strconv.FormatFloat(p.Area, 'f', -1, 64),
		strconv.Itoa(p.WalkMinutes),
		p.URL,
	}
}

// FindNewProperties returns properties that exist in current but not in previous.
// Comparison is based on property ID.
func FindNewProperties(current, previous []Property) []Property {
	prevIDs := make(map[string]bool)
	for _, p := range previous {
		prevIDs[p.ID] = true
	}

	var newProps []Property
	for _, p := range current {
		if !prevIDs[p.ID] {
			newProps = append(newProps, p)
		}
	}

	return newProps
}

// MergeProperties merges two property lists, removing duplicates by ID.
// Properties from 'current' take precedence over 'previous'.
func MergeProperties(current, previous []Property) []Property {
	seen := make(map[string]bool)
	var result []Property

	// Add current properties first
	for _, p := range current {
		if !seen[p.ID] {
			seen[p.ID] = true
			result = append(result, p)
		}
	}

	// Add previous properties that aren't in current
	for _, p := range previous {
		if !seen[p.ID] {
			seen[p.ID] = true
			result = append(result, p)
		}
	}

	return result
}
