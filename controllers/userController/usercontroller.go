package usercontroller

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/nabilulilalbab/TopsisByme/database"
	"github.com/nabilulilalbab/TopsisByme/helper"
	"github.com/nabilulilalbab/TopsisByme/models"
)

// SignupRequest represents the signup request body
type SignupRequest struct {
	NameLengkap     string `json:"nama_lengkap"     example:"John Doe"`
	Email           string `json:"email"            example:"john@example.com"`
	Password        string `json:"password"         example:"password123"`
	ConfirmPassword string `json:"confirm_password" example:"password123"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email           string `json:"email"            example:"john@example.com"`
	Password        string `json:"password"         example:"password123"`
	ConfirmPassword string `json:"confirm_password" example:"password123"`
}

// Signup godoc
// @Summary User registration
// @Description Register a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body SignupRequest true "User registration data"
// @Success 200 {object} helper.Response
// @Failure 400 {object} helper.Response
// @Router /signup [post]
func Signup(c *gin.Context) {
	// Get Email And PAss off req body
	var userInputBody models.User
	if err := c.ShouldBindJSON(&userInputBody); err != nil {
		log.Println("Error Bind Json USer : " + err.Error())
		c.JSON(http.StatusBadRequest, helper.NewResponse("Failed to Ready Body", nil))
		return
	}
	if userInputBody.Password != userInputBody.ConfirmPassword {
		log.Println("Error Password Not match")
		c.JSON(http.StatusBadRequest, helper.NewResponse("password not match", nil))
		return
	}
	// Hash The Password
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(userInputBody.Password), 10)
	if err != nil {
		log.Println("failed Hash Password :" + err.Error())
		c.JSON(http.StatusBadRequest, helper.NewResponse("failed Hash Password", nil))
	}
	hashConfirmPassword, err := bcrypt.GenerateFromPassword(
		[]byte(userInputBody.ConfirmPassword),
		10,
	)
	if err != nil {
		log.Println("failed Hash Confirm  Password :" + err.Error())
		c.JSON(http.StatusBadRequest, helper.NewResponse("failed Hash Password", nil))
	}
	user := models.User{
		NameLengkap:     userInputBody.NameLengkap,
		Email:           userInputBody.Email,
		Password:        string(hashPassword),
		ConfirmPassword: string(hashConfirmPassword),
	}
	result := database.DB.Create(&user)
	if result.Error != nil {
		log.Println("Failed Create User : " + result.Error.Error())
		c.JSON(http.StatusBadRequest, helper.NewResponse("Failed Create USer", nil))
		return
	}
	c.JSON(http.StatusOK, helper.NewResponse("Succes Create User", nil))
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User login credentials"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} helper.Response
// @Router /login [post]
func Login(c *gin.Context) {
	var userInputBody struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}
	if err := c.ShouldBindJSON(&userInputBody); err != nil {
		log.Println("Failed Request body : " + err.Error())
		c.JSON(http.StatusBadRequest, helper.NewResponse("Failed Request Body", nil))
		return
	}
	var user models.User
	database.DB.First(&user, "email = ?", userInputBody.Email)
	if user.Id == 0 {
		c.JSON(http.StatusBadRequest, helper.NewResponse("Invalid Email Or Password", nil))
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userInputBody.Password)); err != nil {
		log.Println("Error Compare : " + err.Error())
		c.JSON(http.StatusBadRequest, helper.NewResponse("Invalid Email Or Password", nil))
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.Id,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		log.Println("invalid to create Token : ", err.Error())
		c.JSON(http.StatusBadRequest, helper.NewResponse("Invalid To Create Token", nil))
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{})
}

// Validate godoc
// @Summary Validate JWT token
// @Description Check if the current JWT token is valid
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} helper.Response
// @Failure 401 {object} helper.Response
// @Security BearerAuth
// @Router /validate [get]
func Validate(c *gin.Context) {
	c.JSON(http.StatusOK, helper.NewResponse("i'm Logged in", nil))
}

// Logout godoc
// @Summary User logout
// @Description Logout user by clearing the authentication cookie
// @Tags Authentication
// @Accept json
// @Produce json
// @Success 200 {object} helper.Response
// @Security BearerAuth
// @Router /logout [post]
func Logout(c *gin.Context) {
	c.SetCookie("Authorization", "", -1, "", "", false, true)
	c.JSON(http.StatusOK, helper.NewResponse("Logged out successfully", nil))
}
