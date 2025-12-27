// Package analyzer provides multiple regression analysis for property value assessment.
package analyzer

import (
	"sort"

	"github.com/alp/suumo-hunter/internal/models"
	"github.com/alp/suumo-hunter/internal/notifier"

	"gonum.org/v1/gonum/mat"
)

const (
	// MinSamples is the minimum number of samples required for regression analysis.
	MinSamples = 10

	// BaseFeatureCount is the number of base features (intercept, area, age, floor, walkMinutes).
	BaseFeatureCount = 5
)

// Analyzer performs regression analysis on property data.
type Analyzer struct {
	minSamples int
}

// regressionModel holds the fitted regression coefficients and station mappings.
type regressionModel struct {
	coefficients []float64
	stations     []string // Sorted list of station names (excluding reference station)
	stationIndex map[string]int
}

// NewAnalyzer creates a new Analyzer instance.
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		minSamples: MinSamples,
	}
}

// extractStations extracts unique station names from properties and returns them sorted.
// The first station in the sorted list is used as the reference category (excluded from dummies).
func extractStations(properties []models.Property) []string {
	stationSet := make(map[string]bool)
	for _, p := range properties {
		if p.NearestStation != "" {
			stationSet[p.NearestStation] = true
		}
	}

	stations := make([]string, 0, len(stationSet))
	for station := range stationSet {
		stations = append(stations, station)
	}
	sort.Strings(stations)

	return stations
}

// buildStationIndex creates a mapping from station name to dummy variable index.
// The first station is excluded as the reference category to avoid multicollinearity.
func buildStationIndex(stations []string) ([]string, map[string]int) {
	if len(stations) <= 1 {
		return nil, nil
	}

	// Skip the first station (reference category)
	dummyStations := stations[1:]
	index := make(map[string]int)
	for i, station := range dummyStations {
		index[station] = i
	}

	return dummyStations, index
}

// Analyze performs multiple regression analysis and calculates bargain scores.
// Returns PropertyWithScore for each input property.
// If there are fewer than MinSamples properties, returns properties with "analyzing" label.
func (a *Analyzer) Analyze(properties []models.Property) []notifier.PropertyWithScore {
	result := make([]notifier.PropertyWithScore, len(properties))

	// Check minimum samples
	if len(properties) < a.minSamples {
		// Not enough data for regression, return with analyzing label
		for i, p := range properties {
			result[i] = notifier.PropertyWithScore{
				Property: p,
				Score:    0,
				Label:    notifier.ScoreLabelAnalyzing,
			}
		}
		return result
	}

	// Perform regression analysis
	model, err := a.fitRegression(properties)
	if err != nil {
		// Regression failed, return with analyzing label
		for i, p := range properties {
			result[i] = notifier.PropertyWithScore{
				Property: p,
				Score:    0,
				Label:    notifier.ScoreLabelAnalyzing,
			}
		}
		return result
	}

	// Calculate scores for each property
	for i, p := range properties {
		predicted := a.predict(p, model)
		actual := p.TotalRent()
		score := predicted - actual // Positive = cheaper than expected (bargain)

		result[i] = notifier.PropertyWithScore{
			Property: p,
			Score:    score,
			Label:    notifier.CalculateScoreLabel(score),
		}
	}

	return result
}

// fitRegression performs multiple linear regression.
// Target variable: Total rent (rent + management_fee)
// Features: Area, Age, Floor, WalkMinutes, Station dummy variables
// Returns regressionModel containing coefficients and station mappings.
func (a *Analyzer) fitRegression(properties []models.Property) (*regressionModel, error) {
	n := len(properties)

	// Extract unique stations and build dummy variable mapping
	allStations := extractStations(properties)
	dummyStations, stationIndex := buildStationIndex(allStations)
	numDummies := len(dummyStations)

	// Total features = base features + station dummies
	numFeatures := BaseFeatureCount + numDummies

	// Build feature matrix X (n x numFeatures) with intercept column
	// Columns: [1, area, age, floor, walkMinutes, station_dummy_1, station_dummy_2, ...]
	xData := make([]float64, n*numFeatures)
	yData := make([]float64, n)

	for i, p := range properties {
		offset := i * numFeatures
		xData[offset+0] = 1                      // Intercept
		xData[offset+1] = p.Area                 // Area (m²)
		xData[offset+2] = float64(p.Age)         // Age (years)
		xData[offset+3] = float64(p.Floor)       // Floor
		xData[offset+4] = float64(p.WalkMinutes) // Walk minutes

		// Station dummy variables
		if idx, ok := stationIndex[p.NearestStation]; ok {
			xData[offset+BaseFeatureCount+idx] = 1
		}
		// If station is the reference category or unknown, all dummies remain 0

		yData[i] = p.TotalRent() // Target: total rent
	}

	X := mat.NewDense(n, numFeatures, xData)
	y := mat.NewVecDense(n, yData)

	// Solve using normal equation: β = (X'X)^(-1) X'y
	var XtX mat.Dense
	XtX.Mul(X.T(), X)

	var XtXInv mat.Dense
	err := XtXInv.Inverse(&XtX)
	if err != nil {
		return nil, err
	}

	var Xty mat.VecDense
	Xty.MulVec(X.T(), y)

	var beta mat.VecDense
	beta.MulVec(&XtXInv, &Xty)

	// Extract coefficients
	coefficients := make([]float64, numFeatures)
	for i := 0; i < numFeatures; i++ {
		coefficients[i] = beta.AtVec(i)
	}

	return &regressionModel{
		coefficients: coefficients,
		stations:     dummyStations,
		stationIndex: stationIndex,
	}, nil
}

// predict calculates the predicted rent for a property.
func (a *Analyzer) predict(p models.Property, model *regressionModel) float64 {
	// Base features: intercept, area, age, floor, walkMinutes
	predicted := model.coefficients[0] +
		model.coefficients[1]*p.Area +
		model.coefficients[2]*float64(p.Age) +
		model.coefficients[3]*float64(p.Floor) +
		model.coefficients[4]*float64(p.WalkMinutes)

	// Add station dummy variable contribution
	if idx, ok := model.stationIndex[p.NearestStation]; ok {
		predicted += model.coefficients[BaseFeatureCount+idx]
	}
	// If station is the reference category or unknown, no additional contribution

	return predicted
}

// AnalyzeNewProperties analyzes only new properties using all properties for regression.
// This is useful when you want to calculate scores only for new properties
// but use the full dataset for more accurate regression.
func (a *Analyzer) AnalyzeNewProperties(allProperties, newProperties []models.Property) []notifier.PropertyWithScore {
	result := make([]notifier.PropertyWithScore, len(newProperties))

	// Check minimum samples
	if len(allProperties) < a.minSamples {
		for i, p := range newProperties {
			result[i] = notifier.PropertyWithScore{
				Property: p,
				Score:    0,
				Label:    notifier.ScoreLabelAnalyzing,
			}
		}
		return result
	}

	// Perform regression on all properties
	model, err := a.fitRegression(allProperties)
	if err != nil {
		for i, p := range newProperties {
			result[i] = notifier.PropertyWithScore{
				Property: p,
				Score:    0,
				Label:    notifier.ScoreLabelAnalyzing,
			}
		}
		return result
	}

	// Calculate scores only for new properties
	for i, p := range newProperties {
		predicted := a.predict(p, model)
		actual := p.TotalRent()
		score := predicted - actual

		result[i] = notifier.PropertyWithScore{
			Property: p,
			Score:    score,
			Label:    notifier.CalculateScoreLabel(score),
		}
	}

	return result
}
