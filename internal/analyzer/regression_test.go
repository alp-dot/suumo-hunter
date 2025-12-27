package analyzer

import (
	"math"
	"testing"

	"github.com/alp/suumo-hunter/internal/models"
	"github.com/alp/suumo-hunter/internal/notifier"
)

// generateTestProperties creates test properties with predictable patterns.
// Rent = 50000 + 2000*area + 500*age - 1000*floor - 500*walkMinutes + noise
func generateTestProperties(n int) []models.Property {
	properties := make([]models.Property, n)
	for i := 0; i < n; i++ {
		area := 20.0 + float64(i%10)*3 // 20-47 m²
		age := i % 20                  // 0-19 years
		floor := (i % 5) + 1           // 1-5 floors
		walkMinutes := (i % 15) + 1    // 1-15 minutes

		// Calculate rent based on a simple formula
		rent := 50000.0 + 2000.0*area - 500.0*float64(age) + 1000.0*float64(floor) - 500.0*float64(walkMinutes)
		managementFee := 5000.0

		properties[i] = models.Property{
			ID:            "jnc_" + string(rune('0'+i%10)),
			Name:          "Test Property",
			Address:       "Tokyo",
			Age:           age,
			Floor:         floor,
			Rent:          rent,
			ManagementFee: managementFee,
			Area:          area,
			WalkMinutes:   walkMinutes,
		}
	}
	return properties
}

func TestAnalyzeWithSufficientData(t *testing.T) {
	analyzer := NewAnalyzer()
	properties := generateTestProperties(20)

	results := analyzer.Analyze(properties)

	if len(results) != len(properties) {
		t.Fatalf("Expected %d results, got %d", len(properties), len(results))
	}

	// All results should have labels (not analyzing)
	analyzingCount := 0
	for _, r := range results {
		if r.Label == notifier.ScoreLabelAnalyzing {
			analyzingCount++
		}
	}

	if analyzingCount > 0 {
		t.Errorf("Expected no analyzing labels with sufficient data, got %d", analyzingCount)
	}
}

func TestAnalyzeWithInsufficientData(t *testing.T) {
	analyzer := NewAnalyzer()
	properties := generateTestProperties(5) // Less than MinSamples

	results := analyzer.Analyze(properties)

	if len(results) != len(properties) {
		t.Fatalf("Expected %d results, got %d", len(properties), len(results))
	}

	// All results should have analyzing label
	for i, r := range results {
		if r.Label != notifier.ScoreLabelAnalyzing {
			t.Errorf("Result[%d] should have analyzing label, got %v", i, r.Label)
		}
		if r.Score != 0 {
			t.Errorf("Result[%d] should have score 0, got %f", i, r.Score)
		}
	}
}

func TestAnalyzeScoreLabels(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create properties where we can predict the outcome
	properties := generateTestProperties(15)

	// Add a cheap property (should be bargain)
	cheapProperty := models.Property{
		ID:            "cheap",
		Name:          "Cheap Property",
		Area:          30.0,
		Age:           5,
		Floor:         3,
		WalkMinutes:   5,
		Rent:          50000, // Very cheap for these specs
		ManagementFee: 5000,
	}
	properties = append(properties, cheapProperty)

	// Add an expensive property (should be expensive)
	expensiveProperty := models.Property{
		ID:            "expensive",
		Name:          "Expensive Property",
		Area:          20.0,
		Age:           15,
		Floor:         1,
		WalkMinutes:   10,
		Rent:          150000, // Very expensive for these specs
		ManagementFee: 10000,
	}
	properties = append(properties, expensiveProperty)

	results := analyzer.Analyze(properties)

	// Find the cheap and expensive properties in results
	var cheapResult, expensiveResult notifier.PropertyWithScore
	for _, r := range results {
		if r.Property.ID == "cheap" {
			cheapResult = r
		}
		if r.Property.ID == "expensive" {
			expensiveResult = r
		}
	}

	// Cheap property should have positive score (bargain)
	if cheapResult.Score <= 0 {
		t.Errorf("Cheap property should have positive score, got %f", cheapResult.Score)
	}
	if cheapResult.Label != notifier.ScoreLabelBargain {
		t.Errorf("Cheap property should be bargain, got %v", cheapResult.Label)
	}

	// Expensive property should have negative score
	if expensiveResult.Score >= 0 {
		t.Errorf("Expensive property should have negative score, got %f", expensiveResult.Score)
	}
	if expensiveResult.Label != notifier.ScoreLabelExpensive {
		t.Errorf("Expensive property should be expensive, got %v", expensiveResult.Label)
	}
}

func TestAnalyzeNewProperties(t *testing.T) {
	analyzer := NewAnalyzer()

	allProperties := generateTestProperties(20)

	newProperties := []models.Property{
		{
			ID:            "new1",
			Name:          "New Property 1",
			Area:          25.0,
			Age:           3,
			Floor:         2,
			WalkMinutes:   5,
			Rent:          70000,
			ManagementFee: 5000,
		},
		{
			ID:            "new2",
			Name:          "New Property 2",
			Area:          35.0,
			Age:           10,
			Floor:         4,
			WalkMinutes:   8,
			Rent:          90000,
			ManagementFee: 5000,
		},
	}

	results := analyzer.AnalyzeNewProperties(allProperties, newProperties)

	if len(results) != len(newProperties) {
		t.Fatalf("Expected %d results, got %d", len(newProperties), len(results))
	}

	// Results should not be analyzing (we have enough data)
	for i, r := range results {
		if r.Label == notifier.ScoreLabelAnalyzing {
			t.Errorf("Result[%d] should not be analyzing", i)
		}
	}
}

func TestAnalyzeNewPropertiesInsufficientData(t *testing.T) {
	analyzer := NewAnalyzer()

	allProperties := generateTestProperties(5) // Less than MinSamples

	newProperties := []models.Property{
		{
			ID:            "new1",
			Name:          "New Property",
			Area:          25.0,
			Age:           3,
			Floor:         2,
			WalkMinutes:   5,
			Rent:          70000,
			ManagementFee: 5000,
		},
	}

	results := analyzer.AnalyzeNewProperties(allProperties, newProperties)

	// All results should have analyzing label
	for i, r := range results {
		if r.Label != notifier.ScoreLabelAnalyzing {
			t.Errorf("Result[%d] should have analyzing label, got %v", i, r.Label)
		}
	}
}

func TestRegressionCoefficients(t *testing.T) {
	analyzer := NewAnalyzer()

	// Create properties with varying features to avoid singular matrix
	properties := make([]models.Property, 20)
	for i := 0; i < 20; i++ {
		area := 20.0 + float64(i)*2
		age := i % 10
		floor := (i % 5) + 1
		walkMinutes := (i % 8) + 3

		// rent = 50000 + 2000*area - 300*age + 500*floor - 200*walkMinutes
		rent := 50000 + 2000*area - 300*float64(age) + 500*float64(floor) - 200*float64(walkMinutes)

		properties[i] = models.Property{
			ID:            "test",
			Area:          area,
			Age:           age,
			Floor:         floor,
			WalkMinutes:   walkMinutes,
			Rent:          rent,
			ManagementFee: 0,
		}
	}

	model, err := analyzer.fitRegression(properties)
	if err != nil {
		t.Fatalf("fitRegression failed: %v", err)
	}

	// Check that area coefficient is close to 2000
	areaCoef := model.coefficients[1]
	if math.Abs(areaCoef-2000) > 100 {
		t.Errorf("Area coefficient should be ~2000, got %f", areaCoef)
	}
}

func TestPrediction(t *testing.T) {
	analyzer := NewAnalyzer()
	model := &regressionModel{
		coefficients: []float64{50000, 2000, -500, 1000, -500},
		stations:     nil,
		stationIndex: nil,
	}

	property := models.Property{
		Area:        25.0,
		Age:         5,
		Floor:       3,
		WalkMinutes: 10,
	}

	predicted := analyzer.predict(property, model)

	// Expected: 50000 + 2000*25 - 500*5 + 1000*3 - 500*10
	// = 50000 + 50000 - 2500 + 3000 - 5000 = 95500
	expected := 95500.0

	if math.Abs(predicted-expected) > 0.01 {
		t.Errorf("Predicted %f, expected %f", predicted, expected)
	}
}

func TestAnalyzeEmptyInput(t *testing.T) {
	analyzer := NewAnalyzer()

	results := analyzer.Analyze([]models.Property{})

	if len(results) != 0 {
		t.Errorf("Expected empty results for empty input, got %d", len(results))
	}
}
