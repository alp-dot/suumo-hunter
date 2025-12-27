// Package analyzer provides multiple regression analysis for property value assessment.
package analyzer

import (
	"github.com/alp/suumo-hunter-go/internal/models"
	"github.com/alp/suumo-hunter-go/internal/notifier"

	"gonum.org/v1/gonum/mat"
)

const (
	// MinSamples is the minimum number of samples required for regression analysis.
	MinSamples = 10
)

// Analyzer performs regression analysis on property data.
type Analyzer struct {
	minSamples int
}

// NewAnalyzer creates a new Analyzer instance.
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		minSamples: MinSamples,
	}
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
	coefficients, err := a.fitRegression(properties)
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
		predicted := a.predict(p, coefficients)
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
// Features: Area, Age, Floor, WalkMinutes
// Returns coefficients [intercept, area, age, floor, walkMinutes]
func (a *Analyzer) fitRegression(properties []models.Property) ([]float64, error) {
	n := len(properties)

	// Build feature matrix X (n x 5) with intercept column
	// Columns: [1, area, age, floor, walkMinutes]
	xData := make([]float64, n*5)
	yData := make([]float64, n)

	for i, p := range properties {
		xData[i*5+0] = 1                      // Intercept
		xData[i*5+1] = p.Area                 // Area (m²)
		xData[i*5+2] = float64(p.Age)         // Age (years)
		xData[i*5+3] = float64(p.Floor)       // Floor
		xData[i*5+4] = float64(p.WalkMinutes) // Walk minutes
		yData[i] = p.TotalRent()              // Target: total rent
	}

	X := mat.NewDense(n, 5, xData)
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
	coefficients := make([]float64, 5)
	for i := 0; i < 5; i++ {
		coefficients[i] = beta.AtVec(i)
	}

	return coefficients, nil
}

// predict calculates the predicted rent for a property.
func (a *Analyzer) predict(p models.Property, coefficients []float64) float64 {
	// coefficients: [intercept, area, age, floor, walkMinutes]
	return coefficients[0] +
		coefficients[1]*p.Area +
		coefficients[2]*float64(p.Age) +
		coefficients[3]*float64(p.Floor) +
		coefficients[4]*float64(p.WalkMinutes)
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
	coefficients, err := a.fitRegression(allProperties)
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
		predicted := a.predict(p, coefficients)
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
