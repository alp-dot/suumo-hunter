// Package models defines data structures for property information.
package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Property represents a rental property listing from SUUMO.
type Property struct {
	ID            string  `csv:"id"`             // 物件ID (jnc_XXXXXXXXXXXX形式)
	Name          string  `csv:"name"`           // 物件名
	Address       string  `csv:"address"`        // 住所
	Age           int     `csv:"age"`            // 築年数
	Floor         int     `csv:"floor"`          // 階数
	Rent          float64 `csv:"rent"`           // 家賃（円）
	ManagementFee float64 `csv:"management_fee"` // 管理費（円）
	Deposit       string  `csv:"deposit"`        // 敷金
	KeyMoney      string  `csv:"key_money"`      // 礼金
	Layout        string  `csv:"layout"`         // 間取り
	Area          float64 `csv:"area"`           // 専有面積（m²）
	WalkMinutes   int     `csv:"walk_minutes"`   // 駅徒歩分数
	URL           string  `csv:"url"`            // 物件詳細URL
}

// TotalRent returns the total monthly cost (rent + management fee).
func (p Property) TotalRent() float64 {
	return p.Rent + p.ManagementFee
}

// TotalRentMan returns the total monthly cost in 万円 (man-yen) unit.
func (p Property) TotalRentMan() float64 {
	return p.TotalRent() / 10000
}

// UniqueKey generates a unique identifier based on property attributes.
// This is used to detect duplicate properties that may have different IDs
// but represent the same physical unit.
// Format: "address|area|layout"
func (p Property) UniqueKey() string {
	return fmt.Sprintf("%s|%.2f|%s", p.Address, p.Area, p.Layout)
}

// Regular expressions for parsing property data.
var (
	// rentManRegex matches patterns like "7.9万円", "10万円", "7.9万", "10万"
	rentManRegex = regexp.MustCompile(`([\d.]+)\s*万円?`)

	// rentYenRegex matches patterns like "5000円", "10000円" (without 万)
	rentYenRegex = regexp.MustCompile(`^([\d,]+)\s*円$`)

	// areaRegex matches patterns like "25.5m²", "25.5㎡", "25.5"
	areaRegex = regexp.MustCompile(`([\d.]+)\s*[m㎡²]*`)

	// ageRegex matches patterns like "築5年", "築15年", "新築"
	ageRegex = regexp.MustCompile(`築(\d+)年`)

	// walkRegex matches patterns like "歩8分", "徒歩8分", "8分"
	walkRegex = regexp.MustCompile(`(?:歩|徒歩)?(\d+)分`)

	// floorRegex matches patterns like "3階", "3-4階" (takes first number)
	// Does not match basement floors like "B1階"
	floorRegex = regexp.MustCompile(`^(\d+)(?:-\d+)?階`)
)

// ParseRent converts a rent string to yen.
// Supports formats like:
//   - "7.9万円", "10万円" (万円 format) -> 79000.0, 100000.0
//   - "5000円", "10,000円" (円 format without 万) -> 5000.0, 10000.0
//
// Returns 0 if the string cannot be parsed.
func ParseRent(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0, nil
	}

	// First try 万円 format
	matches := rentManRegex.FindStringSubmatch(s)
	if len(matches) >= 2 {
		value, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse rent value: %w", err)
		}
		// Convert from 万円 to 円
		return value * 10000, nil
	}

	// Try 円 format (without 万)
	matches = rentYenRegex.FindStringSubmatch(s)
	if len(matches) >= 2 {
		// Remove commas from the number
		numStr := strings.ReplaceAll(matches[1], ",", "")
		value, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse rent value: %w", err)
		}
		return value, nil
	}

	return 0, fmt.Errorf("invalid rent format: %q", s)
}

// ParseArea converts an area string like "25.5m²" to float64 (25.5).
// Returns 0 if the string cannot be parsed.
func ParseArea(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0, nil
	}

	matches := areaRegex.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid area format: %q", s)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse area value: %w", err)
	}

	return value, nil
}

// ParseAge converts an age string like "築5年" to int (5).
// "新築" returns 0.
// Returns 0 if the string cannot be parsed.
func ParseAge(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0, nil
	}

	// Handle "新築" (new construction)
	if strings.Contains(s, "新築") {
		return 0, nil
	}

	matches := ageRegex.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid age format: %q", s)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse age value: %w", err)
	}

	return value, nil
}

// ParseWalkMinutes converts a walking time string like "歩8分" to int (8).
// Returns 0 if the string cannot be parsed.
func ParseWalkMinutes(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0, nil
	}

	matches := walkRegex.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid walk minutes format: %q", s)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse walk minutes value: %w", err)
	}

	return value, nil
}

// ParseFloor converts a floor string like "3階" to int (3).
// For ranges like "3-4階", returns the first number.
// Returns 0 if the string cannot be parsed.
func ParseFloor(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0, nil
	}

	matches := floorRegex.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, fmt.Errorf("invalid floor format: %q", s)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, fmt.Errorf("failed to parse floor value: %w", err)
	}

	return value, nil
}

// ExtractPropertyID extracts the property ID from a SUUMO URL.
// Example: "/chintai/jnc_000102396492/" -> "jnc_000102396492"
func ExtractPropertyID(url string) string {
	// Look for jnc_XXXXXXXXXXXX pattern
	idRegex := regexp.MustCompile(`(jnc_\d+)`)
	matches := idRegex.FindStringSubmatch(url)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}
