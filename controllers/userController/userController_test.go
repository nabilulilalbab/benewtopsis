package usercontroller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	router.POST("/api/signup", Signup)
	router.POST("/api/login", Login)
	router.POST("/api/logout", Logout)
	router.GET("/api/validate", Validate)

	return router
}

func TestSignup(t *testing.T) {
	// Setup test database and router
	db := setupTestDB(t)
	router := setupTestRouter(db)

	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Successfully signup new user",
			requestBody: map[string]string{
				"nama_lengkap":     "Test User",
				"email":            "test@example.com",
				"password":         "password123",
				"confirm_password": "password123",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Succes Create User", response["message"])

				// Verify user was created in database
				var user models.User
				err = db.Where("email = ?", "test@example.com").First(&user).Error
				assert.NoError(t, err)
				assert.Equal(t, "Test User", user.NameLengkap)
			},
		},
		{
			name: "Signup with existing email",
			requestBody: map[string]string{
				"nama_lengkap":     "Test User 2",
				"email":            "test@example.com", // Same email as above
				"password":         "password123",
				"confirm_password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "email already exists")
			},
		},
		{
			name: "Signup with mismatched passwords",
			requestBody: map[string]string{
				"nama_lengkap":     "Test User 3",
				"email":            "test3@example.com",
				"password":         "password123",
				"confirm_password": "different_password",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "passwords do not match")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/signup", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

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

func TestLogin(t *testing.T) {
	// Setup test database and router
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Create test user
	user := models.User{
		NameLengkap:     "Test User",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Successfully login",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["token"])
			},
		},
		{
			name: "Login with wrong password",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "wrong_password",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid credentials")
			},
		},
		{
			name: "Login with non-existent email",
			requestBody: map[string]string{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid credentials")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			reqBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

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

func TestLogout(t *testing.T) {
	// Setup test database and router
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Create test user and get token
	user := models.User{
		NameLengkap:     "Test User",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	if err := db.Create(&user).Error; err != nil {
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

	var loginResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}
	token := loginResponse["token"].(string)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Successfully logout",
			token:          token,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "Logged out successfully", response["message"])
			},
		},
		{
			name:           "Logout without token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Unauthorized")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

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

func TestValidate(t *testing.T) {
	// Setup test database and router
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Create test user and get token
	user := models.User{
		NameLengkap:     "Test User",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	if err := db.Create(&user).Error; err != nil {
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

	var loginResponse map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}
	token := loginResponse["token"].(string)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Successfully validate token",
			token:          token,
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "User is authenticated", response["message"])
			},
		},
		{
			name:           "Validate without token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "Unauthorized")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/validate", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

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
