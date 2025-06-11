package helperTopsis

const (
	Benefit = "benefit"
	Cost    = "cost"
)

type Criterion struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
	Type   string  `json:"type"`
}

type Alternative struct {
	Name   string             `json:"name"`
	Values map[string]float64 `json:"values"`
}

type TOPSISRequest struct {
	Criteria     []Criterion   `json:"criteria"`
	Alternatives []Alternative `json:"alternatives"`
}

type TOPSISResult struct {
	Name             string             `json:"name"`
	ClosenessValue   float64            `json:"closenessvalue"`
	Rank             int                `json:"rank"`
	PositiveDistance float64            `json:"positivedistance"`
	NegativeDistance float64            `json:"negativedistance"`
	NormalizedValues map[string]float64 `json:"normalizedvalues"`
	WeightedValues   map[string]float64 `json:"WeightedValues"`
}

type TOPSISResponse struct {
	Results              []TOPSISResult     `json:"results"`
	IdealPositive        map[string]float64 `json:"idealPositive"`
	IdealNegative        map[string]float64 `json:"idealNegative"`
	NormalizationFactors map[string]float64 `json:"normalizationFactors"`
}
