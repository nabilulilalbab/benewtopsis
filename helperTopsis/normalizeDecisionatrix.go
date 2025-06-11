package helperTopsis

func NormalizeDecisionatrix(
	req TOPSISRequest,
	normFactors map[string]float64,
) map[string]map[string]float64 {
	/*
		🔹 normFactors[criterion.Name] adalah nilai hasil akar kuadrat yang dihitung sebelumnya (dari calculateNormalizationFactors).

		🔹 Normalisasi dilakukan dengan cara membagi nilai asli dengan faktor normalisasi.

		Contoh:

		IPK A = 3.5

		√Σx²(IPK) = 6.077

		Maka: normalized["A"]["IPK"] = 3.5 / 6.077 ≈ 0.576
	*/
	normalized := make(map[string]map[string]float64)
	for _, alt := range req.Alternatives {
		normalized[alt.Name] = make(map[string]float64)
		for _, criterion := range req.Criteria {
			if normFactors[criterion.Name] > 0 {
				normalized[alt.Name][criterion.Name] = alt.Values[criterion.Name] / normFactors[criterion.Name]
			} else {
				normalized[alt.Name][criterion.Name] = 0
			}
		}
	}
	return normalized
}
