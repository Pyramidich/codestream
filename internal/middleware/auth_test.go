package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ilya/codestream/internal/middleware"
)

func TestAuthMiddlewareWithoutToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.AuthMiddleware("secret"))
	router.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.GetString("userID")})
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/me", nil)
	require.NoError(t, err)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddlewareWithValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New().String()
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret"))
	require.NoError(t, err)

	router := gin.New()
	router.Use(middleware.AuthMiddleware("secret"))
	router.GET("/me", func(c *gin.Context) {
		id, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": id})
	})

	w := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodGet, "/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), userID)
}
