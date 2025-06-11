package middleware

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/nabilulilalbab/TopsisByme/database"
	"github.com/nabilulilalbab/TopsisByme/helper"
	"github.com/nabilulilalbab/TopsisByme/models"
)

func RequireAuth(c *gin.Context) {
	// Get Cookie off req
	tokenString, err := c.Cookie("Authorization")
	// Decode / validate it
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusUnauthorized,
			helper.NewResponse("failed Unauthorized", nil),
		)
		return
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("SECRET")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		log.Fatal(err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check The Exp
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				helper.NewResponse("failed Unauthorized", nil),
			)
			return
		}

		// Find The User with Token sub
		var user models.User
		database.DB.First(&user, claims["sub"])

		if user.Id == 0 {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				helper.NewResponse("failed Unauthorized", nil),
			)
			return
		}

		// Attach To request
		c.Set("user", user)
		// continue
		c.Next()
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, helper.NewResponse("failed Unauthorized", nil))
		return
	}
}
