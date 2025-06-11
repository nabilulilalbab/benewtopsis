package helperTopsis

import (
	"fmt"
	"math"
)

func ValidateInput(req TOPSISRequest) error {
	if len(req.Criteria) == 0 {
		return fmt.Errorf("No criteria Provided")
	}
	if len(req.Alternatives) == 0 {
		return fmt.Errorf("No Alternative Provided")
	}
	// check if all weight sum = 1.0
	var weightSum float64
	for _, criterion := range req.Criteria {
		if criterion.Weight < 0 {
			return fmt.Errorf("criterion %s has negative weight", criterion.Name)
		}
		weightSum += criterion.Weight
	}
	// validasi agar jumlah weight nya tetap 1 dan mentoleransi ketika kurang dari 0.0001 , contohnya 0.00001
	if math.Abs(weightSum-1.0) > 0.0001 {
		return fmt.Errorf("weights do not sum to 1.0 (sum: %f)", weightSum)
	}
	criteriaNames := make(map[string]bool)
	for _, criterion := range req.Criteria {
		criteriaNames[criterion.Name] = true
		if criterion.Type != Cost && criterion.Type != Benefit {
			return fmt.Errorf("Invalid Criterion Type for %s : %s", criterion.Name, criterion.Type)
		}
	}
	for _, alt := range req.Alternatives {
		for criterionName := range criteriaNames {
			//  fitur khusus di Go, yaitu multi-value return dari map look value, exists := map[key] , exists berisi boolean
			if _, exists := alt.Values[criterionName]; !exists {
				return fmt.Errorf(
					"Alternative %s is missing Value for criteria %s",
					alt.Name,
					criterionName,
				)
			}
		}
	}
	return nil
}
