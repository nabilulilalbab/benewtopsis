package helperTopsis

func CalculateClosenessAndRank(
	alternatives []Alternative,
	positiveDistances, negativeDistances map[string]float64,
	noramalizeMatrix, weightedMatrix map[string]map[string]float64,
) []TOPSISResult {
	results := make([]TOPSISResult, 0, len(alternatives))

	for _, alt := range alternatives {
		positiveDistance := positiveDistances[alt.Name]
		negativeDistance := negativeDistances[alt.Name]
		closenesValue := 0.0
		if (positiveDistance + negativeDistance) > 0 {
			closenesValue = negativeDistance / (positiveDistance + negativeDistance)
		}

		result := TOPSISResult{
			Name:             alt.Name,
			ClosenessValue:   closenesValue,
			PositiveDistance: positiveDistance,
			NegativeDistance: negativeDistance,
			NormalizedValues: noramalizeMatrix[alt.Name],
			WeightedValues:   weightedMatrix[alt.Name],
		}
		results = append(results, result)
	}
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].ClosenessValue < results[j].ClosenessValue {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	for i := range results {
		results[i].Rank = i + 1
	}
	return results
}
