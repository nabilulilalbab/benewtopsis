// models/topsisResult.go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// RawTopsisData represents the raw input data for TOPSIS calculation
type RawTopsisData struct {
	Alternatives []string          `json:"alternatives"`
	Criteria     map[string]string `json:"criteria"` // key: criteria name, value: criteria type (benefit/cost)
	Values       [][]float64       `json:"values"`   // matrix of values
	Weights      []float64         `json:"weights"`  // weights for each criteria
}

// Value implements the driver.Valuer interface
func (r RawTopsisData) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan implements the sql.Scanner interface
func (r *RawTopsisData) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, &r)
}

// TopsisCalculation represents a TOPSIS calculation record
type TopsisCalculation struct {
	ID        uint          `gorm:"primaryKey" json:"id"         example:"1"`
	UserID    int64         `                  json:"user_id"    example:"1"` // TAMBAHAN: Field untuk mengasosiasikan dengan user
	Name      string        `                  json:"name"       example:"Product Selection Analysis"`
	RawData   RawTopsisData `gorm:"type:json" json:"raw_data"` // Menyimpan data mentah input
	CreatedAt time.Time     `                  json:"created_at"`
	UpdatedAt time.Time     `                  json:"updated_at"`

	// Relasi
	User           User            `gorm:"foreignKey:UserID"              json:"user,omitempty"`
	IdealSolutions []IdealSolution `gorm:"foreignKey:TopsisCalculationID" json:"ideal_solutions,omitempty"`
	Alternatives   []Alternative   `gorm:"foreignKey:TopsisCalculationID" json:"alternatives,omitempty"`
}

// IdealSolution represents the ideal positive and negative solutions for each criteria
type IdealSolution struct {
	ID                  uint    `gorm:"primaryKey" json:"id"                    example:"1"`
	TopsisCalculationID uint    `                  json:"topsis_calculation_id" example:"1"`
	CriteriaName        string  `                  json:"criteria_name"         example:"Cost"`
	IdealPositive       float64 `                  json:"ideal_positive"        example:"0.8"`
	IdealNegative       float64 `                  json:"ideal_negative"        example:"0.2"`
}

// Alternative represents an alternative in TOPSIS calculation
type Alternative struct {
	ID                  uint    `gorm:"primaryKey" json:"id"                    example:"1"`
	TopsisCalculationID uint    `                  json:"topsis_calculation_id" example:"1"`
	Name                string  `                  json:"name"                  example:"Product A"`
	ClosenessValue      float64 `                  json:"closeness_value"       example:"0.75"`
	Rank                int     `                  json:"rank"                  example:"1"`

	CriteriaValues []CriteriaValue `gorm:"foreignKey:AlternativeID" json:"criteria_values,omitempty"`
}

// CriteriaValue represents the normalized and weighted values for each criteria of an alternative
type CriteriaValue struct {
	ID              uint    `gorm:"primaryKey" json:"id"               example:"1"`
	AlternativeID   uint    `                  json:"alternative_id"   example:"1"`
	CriteriaName    string  `                  json:"criteria_name"    example:"Cost"`
	NormalizedValue float64 `                  json:"normalized_value" example:"0.5"`
	WeightedValue   float64 `                  json:"weighted_value"   example:"0.25"`
}
