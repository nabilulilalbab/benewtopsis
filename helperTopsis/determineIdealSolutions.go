package helperTopsis

func DetermineIdealSolutions(
	weightedMatrix map[string]map[string]float64,
	criteria []Criterion,
) (map[string]float64, map[string]float64) {
	idealPositive := make(map[string]float64)
	idealNegative := make(map[string]float64)

	if len(weightedMatrix) > 0 {
		var firstAltName string
		for name := range weightedMatrix {
			firstAltName = name
			break
		}
		// Mengisi idealPositive , idealNegative nilai awal dari hasil map pertama untuk perbandingan , kenapa harus ribet begini , misal kalo pake data dummy 0 maka idealpositif mungkin bisa tapi ideal negatif yang cost bagaimana? karena cost akan memilih yang paling kecil dan 0 akan otomatis terpilih , makanya butuh assign alternatif satu
		// kenapa harus susah2 looping untuk ambil satu nama ? karena map di golang bersifat unordered Karena Go tidak punya cara langsung untuk ambil elemen pertama dari map, karena: Map di Go bersifat unordered. Satu-satunya cara adalah range lalu break seperti di atas.
		for criterionName, value := range weightedMatrix[firstAltName] {
			idealPositive[criterionName] = value
			idealNegative[criterionName] = value
		}
	}
	criteraTypes := make(map[string]string)

	for _, criterion := range criteria {
		criteraTypes[criterion.Name] = criterion.Type
	}
	/*
		criteriaTypes = {
			"IPK": "Benefit",
			"TransportCost": "Cost",
			"Skill": "Benefit",
		}
	*/
	for creationName, criteriaType := range criteraTypes {
		for _, weightedValues := range weightedMatrix {
			value := weightedValues[creationName]
			if criteriaType == Benefit {
				if value > idealPositive[creationName] {
					idealPositive[creationName] = value
				}
				if value < idealNegative[creationName] {
					idealNegative[creationName] = value
				}
			} else {
				if value < idealPositive[creationName] {
					idealPositive[creationName] = value
				}
				if value > idealNegative[creationName] {
					idealNegative[creationName] = value
				}
			}
		}
	}
	return idealPositive, idealNegative
}
