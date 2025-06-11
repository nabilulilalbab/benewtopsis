package helperTopsis

import "math"

func CalculateNormalizationFactors(req TOPSISRequest) map[string]float64 {
	factors := make(map[string]float64)
	for _, criterion := range req.Criteria {
		sumOfSquares := 0.0
		for _, alt := range req.Alternatives {
			value := alt.Values[criterion.Name]
			sumOfSquares += value * value
		}
		factors[criterion.Name] = math.Sqrt(sumOfSquares)
	}
	return factors
	/*
		map[string]float64{
		  "IPK": 6.077,
		  "Skill": 147.39,
		}

		Untuk "IPK":
		go
		Copy
		Edit
		sumOfSquares = 3.5² + 3.2² + 3.8²
		             = 12.25 + 10.24 + 14.44 = 36.93

		normalizationFactor = sqrt(36.93) ≈ 6.077
		Untuk "Skill":
		go
		Copy
		Edit
		sumOfSquares = 80² + 90² + 85²
		             = 6400 + 8100 + 7225 = 21,725

		normalizationFactor = sqrt(21725) ≈ 147.39
	*/
}
