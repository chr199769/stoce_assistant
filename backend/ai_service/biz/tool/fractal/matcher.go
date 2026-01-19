package fractal

import (
	"math"
)

type MatchResult struct {
	StockCode  string
	StartDate  string
	EndDate    string
	Similarity float64 // 0 to 1
	NextDays   []float64
}

// CalculatePearsonCorrelation calculates the Pearson correlation coefficient between two series
func CalculatePearsonCorrelation(x, y []float64) float64 {
	n := float64(len(x))
	if n != float64(len(y)) || n == 0 {
		return 0
	}

	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// NormalizeSeries normalizes a series to start at 1.0 (or 0% change)
func NormalizeSeries(data []float64) []float64 {
	if len(data) == 0 {
		return nil
	}
	base := data[0]
	if base == 0 {
		return data // Avoid division by zero
	}
	normalized := make([]float64, len(data))
	for i, v := range data {
		normalized[i] = v / base
	}
	return normalized
}
