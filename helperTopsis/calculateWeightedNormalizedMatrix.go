package helperTopsis

func CalculateWeightedNormalizedMatrix(
	normalizedMatrix map[string]map[string]float64,
	criteria []Criterion,
) map[string]map[string]float64 {
	weighted := make(map[string]map[string]float64)

	// mengambil weight critera
	criteriaWeights := make(map[string]float64)
	for _, criterion := range criteria {
		criteriaWeights[criterion.Name] = criterion.Weight
	}

	for altName, normalizedValues := range normalizedMatrix {
		weighted[altName] = make(map[string]float64)
		for criterionName, normalizedValue := range normalizedValues {
			weighted[altName][criterionName] = normalizedValue * criteriaWeights[criterionName]
		}
	}
	return weighted
}
