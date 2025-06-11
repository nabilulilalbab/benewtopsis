package topsiscontroller

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	topsis "github.com/nabilulilalbab/TopsisByme/TOPSIS"
	"github.com/nabilulilalbab/TopsisByme/helper"
	"github.com/nabilulilalbab/TopsisByme/helperTopsis"
	"github.com/nabilulilalbab/TopsisByme/models"
)

// HandleTopsis godoc
// @Summary Execute TOPSIS calculation
// @Description Perform TOPSIS (Technique for Order Preference by Similarity to Ideal Solution) calculation
// @Tags TOPSIS
// @Accept json
// @Produce json
// @Param topsis body helperTopsis.TOPSISRequest true "TOPSIS calculation request"
// @Success 200 {object} helper.Response
// @Failure 400 {object} helper.Response
// @Security BearerAuth
// @Router /topsis [post]
func HandleTopsis(c *gin.Context) {
	var req helperTopsis.TOPSISRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error shouldBinjson RequestTopsis : %v", err.Error())
		c.JSON(http.StatusBadRequest, helper.NewResponse("Failed Request Body", nil))
		return
	}
	response, err := topsis.Topsis(req)
	if err != nil {
		log.Printf("Error shouldBinjson RequestTopsis : %v", err.Error())
		c.JSON(http.StatusBadRequest, helper.NewResponse("Failed Calculation Topsis", nil))
		return
	}
	c.JSON(http.StatusOK, helper.NewResponse("Succes Calculation Topsis", response))
}

type SaveTopsisRequest struct {
	Name string `json:"name" example:"My TOPSIS Analysis"`
	Data struct {
		IdealPositive map[string]float64 `json:"idealPositive" example:"map[cost:0.2 quality:0.8]"`
		IdealNegative map[string]float64 `json:"idealNegative" example:"map[cost:0.8 quality:0.2]"`
		Results       []struct {
			Name             string             `json:"name" example:"Alternative 1"`
			ClosenessValue   float64            `json:"closenessvalue" example:"0.75"`
			Rank             int                `json:"rank" example:"1"`
			NormalizedValues map[string]float64 `json:"normalizedvalues" example:"map[cost:0.5 quality:0.8]"`
			WeightedValues   map[string]float64 `json:"WeightedValues" example:"map[cost:0.1 quality:0.4]"`
		} `json:"results"`
	} `json:"data"`
	RawInput struct {
		Alternatives []string          `json:"alternatives"`
		Criteria     map[string]string `json:"criteria"`
		Values       [][]float64       `json:"values"`
		Weights      []float64         `json:"weights"`
	} `json:"raw_input"`
}

// Helper function to get authenticated user
func getUserFromContext(c *gin.Context) (*models.User, bool) {
	userInterface, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	user, ok := userInterface.(models.User)
	if !ok {
		return nil, false
	}
	return &user, true
}

// Helper function to get database connection
func getDatabaseFromContext(c *gin.Context) (*gorm.DB, bool) {
	db, exists := c.Get("db")
	if !exists {
		return nil, false
	}
	gormDB, ok := db.(*gorm.DB)
	if !ok {
		return nil, false
	}
	return gormDB, true
}

// SaveTopsisResult godoc
// @Summary Save TOPSIS calculation result
// @Description Save the result of TOPSIS calculation to database
// @Tags TOPSIS
// @Accept json
// @Produce json
// @Param topsis body SaveTopsisRequest true "TOPSIS result to save"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /topsis/save [post]
func SaveTopsisResult(c *gin.Context) {
	var req SaveTopsisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Get authenticated user
	user, exists := getUserFromContext(c)
	if !exists {
		log.Println("User not found in context")
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Validasi jumlah alternatif
	if len(req.Data.Results) == 0 {
		log.Println("No alternatives provided in results")
		c.JSON(400, gin.H{"error": "Results cannot be empty"})
		return
	}

	// Get database connection
	gormDB, exists := getDatabaseFromContext(c)
	if !exists {
		log.Println("Database connection not found in context")
		c.JSON(500, gin.H{"error": "Database connection not available"})
		return
	}

	// Transaksi database
	tx := gormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic recovered: %v", r)
			c.JSON(500, gin.H{"error": "Internal server error"})
		}
	}()

	// Simpan TopsisCalculation dengan UserID yang benar
	calc := models.TopsisCalculation{
		UserID: user.Id,
		Name:   req.Name,
		RawData: models.RawTopsisData{
			Alternatives: req.RawInput.Alternatives,
			Criteria:     req.RawInput.Criteria,
			Values:       req.RawInput.Values,
			Weights:      req.RawInput.Weights,
		},
	}

	if err := tx.Create(&calc).Error; err != nil {
		tx.Rollback()
		log.Printf("Error saving calculation: %v", err)
		c.JSON(500, gin.H{"error": "Failed to save calculation"})
		return
	}

	log.Printf("Saved calculation with ID: %d for user ID: %d", calc.ID, user.Id)

	// Simpan IdealSolution
	for critName, val := range req.Data.IdealPositive {
		ideal := models.IdealSolution{
			TopsisCalculationID: calc.ID,
			CriteriaName:        critName,
			IdealPositive:       val,
			IdealNegative:       req.Data.IdealNegative[critName],
		}
		if err := tx.Create(&ideal).Error; err != nil {
			tx.Rollback()
			log.Printf("Error saving ideal solution for %s: %v", critName, err)
			c.JSON(500, gin.H{"error": "Failed to save ideal solutions"})
			return
		}
	}

	// Simpan Alternative dan CriteriaValue
	for _, res := range req.Data.Results {
		alt := models.Alternative{
			TopsisCalculationID: calc.ID,
			Name:                res.Name,
			ClosenessValue:      res.ClosenessValue,
			Rank:                res.Rank,
		}
		if err := tx.Create(&alt).Error; err != nil {
			tx.Rollback()
			log.Printf("Error saving alternative %s: %v", res.Name, err)
			c.JSON(500, gin.H{"error": "Failed to save alternative"})
			return
		}

		for critName, normVal := range res.NormalizedValues {
			critVal := models.CriteriaValue{
				AlternativeID:   alt.ID,
				CriteriaName:    critName,
				NormalizedValue: normVal,
				WeightedValue:   res.WeightedValues[critName],
			}
			if err := tx.Create(&critVal).Error; err != nil {
				tx.Rollback()
				log.Printf("Error saving criteria value: %v", err)
				c.JSON(500, gin.H{"error": "Failed to save criteria value"})
				return
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(500, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(200, gin.H{
		"message":        "Topsis result saved successfully",
		"calculation_id": calc.ID,
	})
}

// GetAllTopsisHistory godoc
// @Summary Get all TOPSIS calculation history
// @Description Retrieve all TOPSIS calculations with their results for current user
// @Tags TOPSIS
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /topsis/history [get]
func GetAllTopsisHistory(c *gin.Context) {
	// Get authenticated user
	user, exists := getUserFromContext(c)
	if !exists {
		log.Println("User not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get database connection
	gormDB, exists := getDatabaseFromContext(c)
	if !exists {
		log.Println("Database connection not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection not available"})
		return
	}

	var history []models.TopsisCalculation

	log.Printf("Fetching TOPSIS history for user ID: %d", user.Id)

	// Query hanya data milik user yang sedang login
	err := gormDB.Where("user_id = ?", user.Id).
		Preload("IdealSolutions").
		Preload("Alternatives", func(db *gorm.DB) *gorm.DB {
			return db.Order("`rank` ASC") // Urutkan berdasarkan rank dengan backticks
		}).
		Preload("Alternatives.CriteriaValues").
		Order("created_at DESC"). // Urutkan berdasarkan tanggal terbaru
		Find(&history).Error
	if err != nil {
		log.Printf("Error fetching topsis history for user %d: %v", user.Id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topsis history"})
		return
	}

	log.Printf("Found %d TOPSIS calculations for user ID: %d", len(history), user.Id)

	// Log untuk debugging
	for i, calc := range history {
		log.Printf("Calculation %d: ID=%d, Name=%s, UserID=%d, Alternatives=%d, IdealSolutions=%d",
			i+1, calc.ID, calc.Name, calc.UserID, len(calc.Alternatives), len(calc.IdealSolutions))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Topsis history fetched successfully",
		"data":    history,
		"count":   len(history),
		"user_id": user.Id, // Untuk debugging
	})
}

// TopsisGetById godoc
// @Summary Get TOPSIS calculation by ID
// @Description Retrieve a specific TOPSIS calculation by its ID (only if owned by current user)
// @Tags TOPSIS
// @Accept json
// @Produce json
// @Param id path int true "Calculation ID"
// @Success 200 {object} helper.Response
// @Failure 400 {object} helper.Response
// @Failure 404 {object} helper.Response
// @Failure 500 {object} helper.Response
// @Security BearerAuth
// @Router /topsis/{id} [get]
func TopsisGetById(c *gin.Context) {
	// Ambil ID dari parameter URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Error converting ID parameter: %v", err)
		c.JSON(http.StatusBadRequest, helper.NewResponse("Invalid ID parameter", nil))
		return
	}

	log.Printf("Fetching TOPSIS calculation with ID: %d", id)

	// Get authenticated user
	user, exists := getUserFromContext(c)
	if !exists {
		log.Println("User not found in context")
		c.JSON(http.StatusUnauthorized, helper.NewResponse("Unauthorized", nil))
		return
	}

	// Get database connection
	gormDB, exists := getDatabaseFromContext(c)
	if !exists {
		log.Println("Database connection not found in context")
		c.JSON(http.StatusInternalServerError, helper.NewResponse("Database Connection Error", nil))
		return
	}

	var calculation models.TopsisCalculation

	log.Printf("Querying calculation ID: %d for user ID: %d", id, user.Id)

	// Query dengan filter user_id dan calculation_id untuk memastikan user hanya bisa akses data miliknya
	err = gormDB.Where("id = ? AND user_id = ?", id, user.Id).
		Preload("IdealSolutions").
		Preload("Alternatives", func(db *gorm.DB) *gorm.DB {
			return db.Order("`rank` ASC") // Urutkan berdasarkan rank dengan backticks
		}).
		Preload("Alternatives.CriteriaValues").
		First(&calculation).Error
	if err != nil {
		log.Printf("Error querying calculation with ID %d for user %d: %v", id, user.Id, err)
		if err == gorm.ErrRecordNotFound {
			c.JSON(
				http.StatusNotFound,
				helper.NewResponse(
					"Calculation not found or you don't have permission to access it",
					nil,
				),
			)
		} else {
			c.JSON(http.StatusInternalServerError, helper.NewResponse("Database error: "+err.Error(), nil))
		}
		return
	}

	// Validasi tambahan untuk memastikan data benar-benar milik user
	if calculation.UserID != user.Id {
		log.Printf("Security warning: User %d attempted to access calculation %d owned by user %d",
			user.Id, calculation.ID, calculation.UserID)
		c.JSON(http.StatusForbidden, helper.NewResponse("Access denied", nil))
		return
	}

	// Log untuk debugging
	log.Printf("Successfully found calculation ID=%d, Name=%s for user ID=%d",
		calculation.ID, calculation.Name, calculation.UserID)
	log.Printf("IdealSolutions count: %d", len(calculation.IdealSolutions))
	log.Printf("Alternatives count: %d", len(calculation.Alternatives))

	for i, alt := range calculation.Alternatives {
		log.Printf("Alternative %d: ID=%d, Name=%s, Rank=%d, CriteriaValues=%d",
			i+1, alt.ID, alt.Name, alt.Rank, len(alt.CriteriaValues))
	}

	// Kembalikan hasil
	c.JSON(http.StatusOK, helper.NewResponse("Topsis Calculation Found", calculation))
}

// UpdateTopsisResult godoc
// @Summary Update TOPSIS calculation result
// @Description Update an existing TOPSIS calculation with new alternatives
// @Tags TOPSIS
// @Accept json
// @Produce json
// @Param id path int true "Calculation ID"
// @Param topsis body UpdateTopsisRequest true "Updated TOPSIS data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /topsis/{id} [put]
func UpdateTopsisResult(c *gin.Context) {
	// Get calculation ID from URL parameter
	calcID := c.Param("id")
	if calcID == "" {
		c.JSON(400, gin.H{"error": "Calculation ID is required"})
		return
	}

	// Get authenticated user
	user, exists := getUserFromContext(c)
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Get database connection
	gormDB, exists := getDatabaseFromContext(c)
	if !exists {
		c.JSON(500, gin.H{"error": "Database connection not available"})
		return
	}

	// Get existing calculation
	var calc models.TopsisCalculation
	if err := gormDB.Where("id = ? AND user_id = ?", calcID, user.Id).First(&calc).Error; err != nil {
		c.JSON(404, gin.H{"error": "Calculation not found"})
		return
	}

	// Parse request body
	var req UpdateTopsisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Validasi jumlah alternatif
	if len(req.Alternatives) == 0 {
		c.JSON(400, gin.H{"error": "Alternatives cannot be empty"})
		return
	}

	// Validasi jumlah nilai per alternatif
	expectedValues := len(calc.RawData.Criteria)
	for i, alt := range req.Alternatives {
		if len(alt.Values) != expectedValues {
			c.JSON(400, gin.H{
				"error": fmt.Sprintf("Alternative %d (%s) must have exactly %d values", i+1, alt.Name, expectedValues),
			})
			return
		}
	}

	// Validasi nama kriteria
	for i, alt := range req.Alternatives {
		for j, val := range alt.Values {
			critNames := make([]string, 0, len(calc.RawData.Criteria))
			for name := range calc.RawData.Criteria {
				critNames = append(critNames, name)
			}
			sort.Strings(critNames)
			critName := critNames[j]
			if val.CriteriaName != critName {
				c.JSON(400, gin.H{
					"error": fmt.Sprintf("Invalid criteria name for alternative %d (%s): expected %s, got %s",
						i+1, alt.Name, critName, val.CriteriaName),
				})
				return
			}
		}
	}

	// Transaksi database
	tx := gormDB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Internal server error"})
		}
	}()

	// Hapus semua criteria_values yang terkait dengan alternatif yang akan dihapus
	if err := tx.Where("alternative_id IN (SELECT id FROM alternatives WHERE topsis_calculation_id = ?)", calc.ID).Delete(&models.CriteriaValue{}).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Failed to delete criteria values"})
		return
	}

	// Hapus semua alternatif yang ada
	if err := tx.Where("topsis_calculation_id = ?", calc.ID).Delete(&models.Alternative{}).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Failed to delete alternatives"})
		return
	}

	// Update raw data
	calc.RawData.Alternatives = make([]string, len(req.Alternatives))
	calc.RawData.Values = make([][]float64, len(req.Alternatives))
	for i, alt := range req.Alternatives {
		calc.RawData.Alternatives[i] = alt.Name
		calc.RawData.Values[i] = make([]float64, len(alt.Values))
		for j, val := range alt.Values {
			calc.RawData.Values[i][j] = val.Value
		}
	}

	// Simpan perubahan raw data
	if err := tx.Save(&calc).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Failed to update raw data"})
		return
	}

	// Konversi raw data ke format TOPSISRequest untuk kalkulasi ulang
	topsisReq := helperTopsis.TOPSISRequest{
		Criteria:     make([]helperTopsis.Criterion, 0, len(calc.RawData.Criteria)),
		Alternatives: make([]helperTopsis.Alternative, 0, len(calc.RawData.Alternatives)),
	}

	// Konversi kriteria
	for name, critType := range calc.RawData.Criteria {
		topsisReq.Criteria = append(topsisReq.Criteria, helperTopsis.Criterion{
			Name:   name,
			Weight: calc.RawData.Weights[len(topsisReq.Criteria)],
			Type:   critType,
		})
	}

	// Konversi alternatif
	for i, altName := range calc.RawData.Alternatives {
		alt := helperTopsis.Alternative{
			Name:   altName,
			Values: make(map[string]float64),
		}
		for j, crit := range topsisReq.Criteria {
			alt.Values[crit.Name] = calc.RawData.Values[i][j]
		}
		topsisReq.Alternatives = append(topsisReq.Alternatives, alt)
	}

	// Kalkulasi ulang TOPSIS
	topsisResult, err := topsis.Topsis(topsisReq)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Failed to recalculate TOPSIS: " + err.Error()})
		return
	}

	// Simpan hasil kalkulasi ulang
	// Hapus ideal solutions yang ada
	if err := tx.Where("topsis_calculation_id = ?", calc.ID).Delete(&models.IdealSolution{}).Error; err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": "Failed to delete existing ideal solutions"})
		return
	}

	// Simpan ideal solutions baru
	for critName, val := range topsisResult.IdealPositive {
		ideal := models.IdealSolution{
			TopsisCalculationID: calc.ID,
			CriteriaName:        critName,
			IdealPositive:       val,
			IdealNegative:       topsisResult.IdealNegative[critName],
		}
		if err := tx.Create(&ideal).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Failed to save ideal solutions"})
			return
		}
	}

	// Simpan alternatif dan nilai kriteria baru
	for _, res := range topsisResult.Results {
		alt := models.Alternative{
			TopsisCalculationID: calc.ID,
			Name:                res.Name,
			ClosenessValue:      res.ClosenessValue,
			Rank:                res.Rank,
		}
		if err := tx.Create(&alt).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "Failed to save alternative"})
			return
		}

		for critName, normVal := range res.NormalizedValues {
			critVal := models.CriteriaValue{
				AlternativeID:   alt.ID,
				CriteriaName:    critName,
				NormalizedValue: normVal,
				WeightedValue:   res.WeightedValues[critName],
			}
			if err := tx.Create(&critVal).Error; err != nil {
				tx.Rollback()
				c.JSON(500, gin.H{"error": "Failed to save criteria value"})
				return
			}
		}
	}

	// Commit transaksi
	if err := tx.Commit().Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to commit transaction"})
		return
	}

	c.JSON(200, gin.H{
		"message":        "Topsis calculation updated successfully",
		"calculation_id": calc.ID,
		"result":         topsisResult,
	})
}

// UpdateTopsisRequest represents the request body for updating a TOPSIS calculation
type UpdateTopsisRequest struct {
	Alternatives []struct {
		Name   string `json:"name" example:"Alternative 1"`
		Values []struct {
			CriteriaName string  `json:"criteria_name" example:"cost"`
			Value        float64 `json:"value" example:"100"`
		} `json:"values"`
	} `json:"alternatives"`
}
