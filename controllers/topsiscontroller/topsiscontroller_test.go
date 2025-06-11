package topsiscontroller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/nabilulilalbab/TopsisByme/database"
	"github.com/nabilulilalbab/TopsisByme/initializers"
	"github.com/nabilulilalbab/TopsisByme/models"
)

// setupTestDB initializes a test database
func setupTestDB(t *testing.T) *gorm.DB {
	initializers.LoadVariables()
	database.InitDB()

	// Clean up existing data
	database.DB.Exec("DELETE FROM criteria_values")
	database.DB.Exec("DELETE FROM alternatives")
	database.DB.Exec("DELETE FROM ideal_solutions")
	database.DB.Exec("DELETE FROM topsis_calculations")
	database.DB.Exec("DELETE FROM users")

	return database.DB
}

// setupTestRouter creates a test router with necessary middleware
func setupTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Middleware untuk inject db ke context
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Add routes
	router.POST("/api/login", func(c *gin.Context) {
		var loginReq struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&loginReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		var user models.User
		if err := database.DB.Where("email = ?", loginReq.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// In a real app, we would verify the password hash
		if loginReq.Password != user.Password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		// Generate a simple token for testing
		token := "test_token_" + user.Email
		c.JSON(http.StatusOK, gin.H{"token": token})
	})

	// Add test middleware to handle authentication
	router.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		// For testing, we'll accept any token that starts with "test_token_"
		if !strings.HasPrefix(token, "test_token_") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Get user from token
		email := strings.TrimPrefix(token, "test_token_")
		var user models.User
		if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Next()
	})

	router.PUT("/api/topsis/:id", UpdateTopsisResult)
	router.POST("/api/topsis/save", SaveTopsisResult)
	return router
}

// createTestUserAndGetToken creates a test user and returns their JWT token
func createTestUserAndGetToken(t *testing.T, router *gin.Engine) string {
	// Create test user
	user := models.User{
		NameLengkap:     "Test User",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Login to get token
	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	reqBody, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}

	token, ok := response["token"].(string)
	if !ok {
		t.Fatalf("Token not found in response: %v", response)
	}
	return token
}

// createTestTopsisCalculation creates a test TOPSIS calculation
func createTestTopsisCalculation(t *testing.T, router *gin.Engine, token string) models.TopsisCalculation {
	// Get user ID from token
	var user models.User
	if err := database.DB.Where("email = ?", "test@example.com").First(&user).Error; err != nil {
		t.Fatalf("Failed to get test user: %v", err)
	}

	// Create initial calculation
	calc := models.TopsisCalculation{
		UserID: user.Id, // Use the actual user ID
		Name:   "Test Calculation",
		RawData: models.RawTopsisData{
			Alternatives: []string{"A1", "A2"},
			Criteria: map[string]string{
				"C1": "benefit",
				"C2": "cost",
			},
			Values: [][]float64{
				{1.0, 2.0},
				{3.0, 4.0},
			},
			Weights: []float64{0.5, 0.5},
		},
	}
	if err := database.DB.Create(&calc).Error; err != nil {
		t.Fatalf("Failed to create test calculation: %v", err)
	}
	return calc
}

type SaveTopsisData struct {
	IdealPositive map[string]float64 `json:"idealPositive"`
	IdealNegative map[string]float64 `json:"idealNegative"`
	Results       []struct {
		Name             string             `json:"name"`
		ClosenessValue   float64            `json:"closenessvalue"`
		Rank             int                `json:"rank"`
		NormalizedValues map[string]float64 `json:"normalizedvalues"`
		WeightedValues   map[string]float64 `json:"WeightedValues"`
	} `json:"results"`
}

type SaveTopsisRequestTest struct {
	Name     string         `json:"name"`
	Data     SaveTopsisData `json:"data"`
	RawInput struct {
		Alternatives []string          `json:"alternatives"`
		Criteria     map[string]string `json:"criteria"`
		Values       [][]float64       `json:"values"`
		Weights      []float64         `json:"weights"`
	} `json:"raw_input"`
}

func TestUpdateTopsisResult(t *testing.T) {
	// Setup test database and router
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Create test user and get token
	token := createTestUserAndGetToken(t, router)

	// Create initial TOPSIS calculation and store its ID
	calc := createTestTopsisCalculation(t, router, token)
	calculationID := calc.ID

	tests := []struct {
		name           string
		calculationID  string
		requestBody    UpdateTopsisRequest
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:          "Successfully update alternatives with recalculation",
			calculationID: fmt.Sprintf("%d", calculationID),
			requestBody: UpdateTopsisRequest{
				Alternatives: []struct {
					Name   string `json:"name" example:"Alternative 1"`
					Values []struct {
						CriteriaName string  `json:"criteria_name" example:"cost"`
						Value        float64 `json:"value" example:"100"`
					} `json:"values"`
				}{
					{
						Name: "A1",
						Values: []struct {
							CriteriaName string  `json:"criteria_name" example:"cost"`
							Value        float64 `json:"value" example:"100"`
						}{
							{CriteriaName: "C1", Value: 1.0},
							{CriteriaName: "C2", Value: 2.0},
						},
					},
					{
						Name: "A2",
						Values: []struct {
							CriteriaName string  `json:"criteria_name" example:"cost"`
							Value        float64 `json:"value" example:"100"`
						}{
							{CriteriaName: "C1", Value: 3.0},
							{CriteriaName: "C2", Value: 4.0},
						},
					},
					{
						Name: "A3",
						Values: []struct {
							CriteriaName string  `json:"criteria_name" example:"cost"`
							Value        float64 `json:"value" example:"100"`
						}{
							{CriteriaName: "C1", Value: 5.0},
							{CriteriaName: "C2", Value: 6.0},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Topsis calculation updated successfully", response["message"])

				// Verify recalculation results
				result, ok := response["result"].(map[string]interface{})
				assert.True(t, ok, "Result should be present in response")

				// Verify ideal solutions
				idealPositive, ok := result["idealPositive"].(map[string]interface{})
				assert.True(t, ok, "Ideal positive solutions should be present")
				assert.NotEmpty(t, idealPositive)

				idealNegative, ok := result["idealNegative"].(map[string]interface{})
				assert.True(t, ok, "Ideal negative solutions should be present")
				assert.NotEmpty(t, idealNegative)

				// Verify results
				results, ok := result["results"].([]interface{})
				assert.True(t, ok, "Results should be present")
				assert.Equal(t, 3, len(results), "Should have 3 alternatives")

				// Verify database state
				var updatedCalc models.TopsisCalculation
				err = db.Preload("IdealSolutions").
					Preload("Alternatives", func(db *gorm.DB) *gorm.DB {
						return db.Order("`rank` ASC")
					}).
					Preload("Alternatives.CriteriaValues").
					First(&updatedCalc, calculationID).Error
				assert.NoError(t, err)

				// Verify alternatives
				assert.Equal(t, 3, len(updatedCalc.Alternatives))

				// Verify ideal solutions
				assert.Equal(t, 2, len(updatedCalc.IdealSolutions))

				// Verify criteria values
				for _, alt := range updatedCalc.Alternatives {
					assert.Equal(t, 2, len(alt.CriteriaValues))
				}
			},
		},
		{
			name:          "Try to modify criteria",
			calculationID: fmt.Sprintf("%d", calculationID),
			requestBody: UpdateTopsisRequest{
				Alternatives: []struct {
					Name   string `json:"name" example:"Alternative 1"`
					Values []struct {
						CriteriaName string  `json:"criteria_name" example:"cost"`
						Value        float64 `json:"value" example:"100"`
					} `json:"values"`
				}{
					{
						Name: "A1",
						Values: []struct {
							CriteriaName string  `json:"criteria_name" example:"cost"`
							Value        float64 `json:"value" example:"100"`
						}{
							{CriteriaName: "C3", Value: 1.0}, // Invalid criteria name
							{CriteriaName: "C2", Value: 2.0},
						},
					},
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid criteria name")
			},
		},
		{
			name:          "Add new alternative with recalculation",
			calculationID: fmt.Sprintf("%d", calculationID),
			requestBody: UpdateTopsisRequest{
				Alternatives: []struct {
					Name   string `json:"name" example:"Alternative 1"`
					Values []struct {
						CriteriaName string  `json:"criteria_name" example:"cost"`
						Value        float64 `json:"value" example:"100"`
					} `json:"values"`
				}{
					{
						Name: "A1",
						Values: []struct {
							CriteriaName string  `json:"criteria_name" example:"cost"`
							Value        float64 `json:"value" example:"100"`
						}{
							{CriteriaName: "C1", Value: 1.0},
							{CriteriaName: "C2", Value: 2.0},
						},
					},
					{
						Name: "A2",
						Values: []struct {
							CriteriaName string  `json:"criteria_name" example:"cost"`
							Value        float64 `json:"value" example:"100"`
						}{
							{CriteriaName: "C1", Value: 3.0},
							{CriteriaName: "C2", Value: 4.0},
						},
					},
					{
						Name: "A3",
						Values: []struct {
							CriteriaName string  `json:"criteria_name" example:"cost"`
							Value        float64 `json:"value" example:"100"`
						}{
							{CriteriaName: "C1", Value: 5.0},
							{CriteriaName: "C2", Value: 6.0},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Topsis calculation updated successfully", response["message"])

				// Verify recalculation results
				result, ok := response["result"].(map[string]interface{})
				assert.True(t, ok, "Result should be present in response")

				// Verify results length
				results, ok := result["results"].([]interface{})
				assert.True(t, ok, "Results should be present")
				assert.Equal(t, 3, len(results), "Should have 3 alternatives")

				// Verify database state
				var updatedCalc models.TopsisCalculation
				err = db.Preload("IdealSolutions").
					Preload("Alternatives", func(db *gorm.DB) *gorm.DB {
						return db.Order("`rank` ASC")
					}).
					Preload("Alternatives.CriteriaValues").
					First(&updatedCalc, calculationID).Error
				assert.NoError(t, err)

				// Verify alternatives
				assert.Equal(t, 3, len(updatedCalc.Alternatives))

				// Verify ideal solutions
				assert.Equal(t, 2, len(updatedCalc.IdealSolutions))

				// Verify criteria values
				for _, alt := range updatedCalc.Alternatives {
					assert.Equal(t, 2, len(alt.CriteriaValues))
				}
			},
		},
		{
			name:          "Remove alternative with recalculation",
			calculationID: fmt.Sprintf("%d", calculationID),
			requestBody: UpdateTopsisRequest{
				Alternatives: []struct {
					Name   string `json:"name" example:"Alternative 1"`
					Values []struct {
						CriteriaName string  `json:"criteria_name" example:"cost"`
						Value        float64 `json:"value" example:"100"`
					} `json:"values"`
				}{
					{
						Name: "A1",
						Values: []struct {
							CriteriaName string  `json:"criteria_name" example:"cost"`
							Value        float64 `json:"value" example:"100"`
						}{
							{CriteriaName: "C1", Value: 1.0},
							{CriteriaName: "C2", Value: 2.0},
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Topsis calculation updated successfully", response["message"])

				// Verify recalculation results
				result, ok := response["result"].(map[string]interface{})
				assert.True(t, ok, "Result should be present in response")

				// Verify results length
				results, ok := result["results"].([]interface{})
				assert.True(t, ok, "Results should be present")
				assert.Equal(t, 1, len(results), "Should have 1 alternative")

				// Verify database state
				var updatedCalc models.TopsisCalculation
				err = db.Preload("IdealSolutions").
					Preload("Alternatives", func(db *gorm.DB) *gorm.DB {
						return db.Order("`rank` ASC")
					}).
					Preload("Alternatives.CriteriaValues").
					First(&updatedCalc, calculationID).Error
				assert.NoError(t, err)

				// Verify alternatives
				assert.Equal(t, 1, len(updatedCalc.Alternatives))

				// Verify ideal solutions
				assert.Equal(t, 2, len(updatedCalc.IdealSolutions))

				// Verify criteria values
				for _, alt := range updatedCalc.Alternatives {
					assert.Equal(t, 2, len(alt.CriteriaValues))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/topsis/%s", tt.calculationID), bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			// Create response recorder
			w := httptest.NewRecorder()

			// Send request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response
			tt.checkResponse(t, w)
		})
	}
}

func TestSaveTopsisResult(t *testing.T) {
	// Setup test database and router
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Create test user and get token
	token := createTestUserAndGetToken(t, router)

	tests := []struct {
		name           string
		requestBody    SaveTopsisRequestTest
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Successfully save TOPSIS result with raw data",
			requestBody: SaveTopsisRequestTest{
				Name: "Test Calculation",
				Data: SaveTopsisData{
					IdealPositive: map[string]float64{
						"C1": 0.8,
						"C2": 0.2,
					},
					IdealNegative: map[string]float64{
						"C1": 0.2,
						"C2": 0.8,
					},
					Results: []struct {
						Name             string             `json:"name"`
						ClosenessValue   float64            `json:"closenessvalue"`
						Rank             int                `json:"rank"`
						NormalizedValues map[string]float64 `json:"normalizedvalues"`
						WeightedValues   map[string]float64 `json:"WeightedValues"`
					}{
						{
							Name:           "A1",
							ClosenessValue: 0.75,
							Rank:           1,
							NormalizedValues: map[string]float64{
								"C1": 0.5,
								"C2": 0.5,
							},
							WeightedValues: map[string]float64{
								"C1": 0.4,
								"C2": 0.25,
							},
						},
					},
				},
				RawInput: struct {
					Alternatives []string          `json:"alternatives"`
					Criteria     map[string]string `json:"criteria"`
					Values       [][]float64       `json:"values"`
					Weights      []float64         `json:"weights"`
				}{
					Alternatives: []string{"A1"},
					Criteria: map[string]string{
						"C1": "benefit",
						"C2": "cost",
					},
					Values:  [][]float64{{1.0, 2.0}},
					Weights: []float64{0.5, 0.5},
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Topsis result saved successfully", response["message"])
				assert.NotNil(t, response["calculation_id"])

				// Verify raw data was saved
				var calc models.TopsisCalculation
				err = db.First(&calc, response["calculation_id"]).Error
				assert.NoError(t, err)
				assert.NotEmpty(t, calc.RawData.Alternatives)
				assert.NotEmpty(t, calc.RawData.Criteria)
				assert.NotEmpty(t, calc.RawData.Values)
				assert.NotEmpty(t, calc.RawData.Weights)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/topsis/save", bytes.NewBuffer(reqBody))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Serve request
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response
			tt.checkResponse(t, w)
		})
	}
}
