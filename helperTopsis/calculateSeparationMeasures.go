package helperTopsis

import "math"

func CalculateSeparationMeasures(
	weightedMatrix map[string]map[string]float64,
	idealPositive, idealNegative map[string]float64,
) (map[string]float64, map[string]float64) {
	positiveDistance := make(map[string]float64)
	negativeDistance := make(map[string]float64)
	for altName, weightedValues := range weightedMatrix {
		positiveSum := 0.0
		negativeSum := 0.0
		for criterionName, value := range weightedValues {
			positiveSquareDiff := math.Pow(value-idealPositive[criterionName], 2)
			negativeSquareDiff := math.Pow(value-idealNegative[criterionName], 2)
			positiveSum += positiveSquareDiff
			negativeSum += negativeSquareDiff
		}

		positiveDistance[altName] = math.Sqrt(positiveSum)
		negativeDistance[altName] = math.Sqrt(negativeSum)
	}
	return positiveDistance, negativeDistance
}
