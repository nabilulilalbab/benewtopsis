package topsis

import (
	"github.com/nabilulilalbab/TopsisByme/helperTopsis"
)

func Topsis(req helperTopsis.TOPSISRequest) (helperTopsis.TOPSISResponse, error) {
	if err := helperTopsis.ValidateInput(req); err != nil {
		return helperTopsis.TOPSISResponse{}, err
	}

	normFaktors := helperTopsis.CalculateNormalizationFactors(req)
	normalizedMatrix := helperTopsis.NormalizeDecisionatrix(req, normFaktors)
	weightedMatrix := helperTopsis.CalculateWeightedNormalizedMatrix(normalizedMatrix, req.Criteria)
	idealPositive, idealNegative := helperTopsis.DetermineIdealSolutions(
		weightedMatrix,
		req.Criteria,
	)
	positiveDistances, negativeDistances := helperTopsis.CalculateSeparationMeasures(
		weightedMatrix,
		idealPositive,
		idealNegative,
	)
	results := helperTopsis.CalculateClosenessAndRank(
		req.Alternatives,
		positiveDistances,
		negativeDistances,
		normalizedMatrix,
		weightedMatrix,
	)
	return helperTopsis.TOPSISResponse{
		Results:              results,
		IdealPositive:        idealPositive,
		IdealNegative:        idealNegative,
		NormalizationFactors: normFaktors,
	}, nil
}
