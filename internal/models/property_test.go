package models

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseRent(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{
			name:  "standard format with 万円",
			input: "7.9万円",
			want:  79000,
		},
		{
			name:  "integer with 万円",
			input: "10万円",
			want:  100000,
		},
		{
			name:  "format with 万 only",
			input: "7.9万",
			want:  79000,
		},
		{
			name:  "with spaces",
			input: " 8.5 万円 ",
			want:  85000,
		},
		{
			name:  "yen format without 万",
			input: "5000円",
			want:  5000,
		},
		{
			name:  "yen format with comma",
			input: "10,000円",
			want:  10000,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
		{
			name:  "dash",
			input: "-",
			want:  0,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRent(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseRent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseArea(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{
			name:  "standard format with m²",
			input: "25.5m²",
			want:  25.5,
		},
		{
			name:  "format with ㎡",
			input: "30.0㎡",
			want:  30.0,
		},
		{
			name:  "integer value",
			input: "25m²",
			want:  25,
		},
		{
			name:  "number only",
			input: "25.5",
			want:  25.5,
		},
		{
			name:  "with spaces",
			input: " 25.5 m² ",
			want:  25.5,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
		{
			name:  "dash",
			input: "-",
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseArea(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArea() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseArea() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAge(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "standard format",
			input: "築5年",
			want:  5,
		},
		{
			name:  "two digit years",
			input: "築15年",
			want:  15,
		},
		{
			name:  "new construction",
			input: "新築",
			want:  0,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
		{
			name:  "dash",
			input: "-",
			want:  0,
		},
		{
			name:    "invalid format",
			input:   "5年",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAge(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseAge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseWalkMinutes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "standard format with 歩",
			input: "歩8分",
			want:  8,
		},
		{
			name:  "format with 徒歩",
			input: "徒歩10分",
			want:  10,
		},
		{
			name:  "number and 分 only",
			input: "8分",
			want:  8,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
		{
			name:  "dash",
			input: "-",
			want:  0,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseWalkMinutes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseWalkMinutes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseWalkMinutes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseFloor(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{
			name:  "standard format",
			input: "3階",
			want:  3,
		},
		{
			name:  "range format",
			input: "3-4階",
			want:  3,
		},
		{
			name:  "two digit floor",
			input: "12階",
			want:  12,
		},
		{
			name:  "empty string",
			input: "",
			want:  0,
		},
		{
			name:  "dash",
			input: "-",
			want:  0,
		},
		{
			name:    "basement format",
			input:   "B1階",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFloor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFloor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFloor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractPropertyID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "full URL path",
			input: "/chintai/jnc_000102396492/",
			want:  "jnc_000102396492",
		},
		{
			name:  "full URL",
			input: "https://suumo.jp/chintai/jnc_000102396492/",
			want:  "jnc_000102396492",
		},
		{
			name:  "ID only",
			input: "jnc_000102396492",
			want:  "jnc_000102396492",
		},
		{
			name:  "no ID",
			input: "/chintai/",
			want:  "",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPropertyID(tt.input)
			if got != tt.want {
				t.Errorf("ExtractPropertyID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPropertyTotalRent(t *testing.T) {
	p := Property{
		Rent:          79000,
		ManagementFee: 5000,
	}

	if got := p.TotalRent(); got != 84000 {
		t.Errorf("TotalRent() = %v, want 84000", got)
	}

	if got := p.TotalRentMan(); got != 8.4 {
		t.Errorf("TotalRentMan() = %v, want 8.4", got)
	}
}

func TestCSVRoundTrip(t *testing.T) {
	original := []Property{
		{
			ID:            "jnc_000102396492",
			Name:          "テストマンション",
			Address:       "東京都渋谷区",
			Age:           5,
			Floor:         3,
			Rent:          79000,
			ManagementFee: 5000,
			Deposit:       "1ヶ月",
			KeyMoney:      "1ヶ月",
			Layout:        "1K",
			Area:          25.5,
			WalkMinutes:   8,
			URL:           "https://suumo.jp/chintai/jnc_000102396492/",
		},
		{
			ID:            "jnc_000102396493",
			Name:          "テストアパート",
			Address:       "東京都新宿区",
			Age:           10,
			Floor:         2,
			Rent:          65000,
			ManagementFee: 3000,
			Deposit:       "-",
			KeyMoney:      "1ヶ月",
			Layout:        "1R",
			Area:          20.0,
			WalkMinutes:   5,
			URL:           "https://suumo.jp/chintai/jnc_000102396493/",
		},
	}

	// Write to CSV
	var buf bytes.Buffer
	if err := SaveToCSV(&buf, original); err != nil {
		t.Fatalf("SaveToCSV() error = %v", err)
	}

	// Read back from CSV
	loaded, err := LoadFromCSV(&buf)
	if err != nil {
		t.Fatalf("LoadFromCSV() error = %v", err)
	}

	// Compare
	if len(loaded) != len(original) {
		t.Fatalf("LoadFromCSV() returned %d properties, want %d", len(loaded), len(original))
	}

	for i := range original {
		if loaded[i].ID != original[i].ID {
			t.Errorf("Property[%d].ID = %v, want %v", i, loaded[i].ID, original[i].ID)
		}
		if loaded[i].Name != original[i].Name {
			t.Errorf("Property[%d].Name = %v, want %v", i, loaded[i].Name, original[i].Name)
		}
		if loaded[i].Rent != original[i].Rent {
			t.Errorf("Property[%d].Rent = %v, want %v", i, loaded[i].Rent, original[i].Rent)
		}
		if loaded[i].Area != original[i].Area {
			t.Errorf("Property[%d].Area = %v, want %v", i, loaded[i].Area, original[i].Area)
		}
	}
}

func TestLoadFromCSVEmpty(t *testing.T) {
	reader := strings.NewReader("")
	props, err := LoadFromCSV(reader)
	if err != nil {
		t.Errorf("LoadFromCSV() with empty input error = %v", err)
	}
	if len(props) != 0 {
		t.Errorf("LoadFromCSV() with empty input returned %d properties, want 0", len(props))
	}
}

func TestFindNewProperties(t *testing.T) {
	previous := []Property{
		{ID: "jnc_001"},
		{ID: "jnc_002"},
	}
	current := []Property{
		{ID: "jnc_002"},
		{ID: "jnc_003"},
		{ID: "jnc_004"},
	}

	newProps := FindNewProperties(current, previous)

	if len(newProps) != 2 {
		t.Fatalf("FindNewProperties() returned %d properties, want 2", len(newProps))
	}

	expectedIDs := map[string]bool{"jnc_003": true, "jnc_004": true}
	for _, p := range newProps {
		if !expectedIDs[p.ID] {
			t.Errorf("FindNewProperties() returned unexpected ID: %s", p.ID)
		}
	}
}

func TestMergeProperties(t *testing.T) {
	previous := []Property{
		{ID: "jnc_001", Name: "Old1"},
		{ID: "jnc_002", Name: "Old2"},
	}
	current := []Property{
		{ID: "jnc_002", Name: "New2"},
		{ID: "jnc_003", Name: "New3"},
	}

	merged := MergeProperties(current, previous)

	if len(merged) != 3 {
		t.Fatalf("MergeProperties() returned %d properties, want 3", len(merged))
	}

	// Check that current properties take precedence
	for _, p := range merged {
		if p.ID == "jnc_002" && p.Name != "New2" {
			t.Errorf("MergeProperties() should prefer current for ID jnc_002, got Name=%s", p.Name)
		}
	}
}
