package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/ksamf/video-upscaling/backend/internal/interfaces"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	user   interfaces.User
	config *config.Config
}

func New(users interfaces.User, conf *config.Config) *Auth {
	return &Auth{
		user:   users,
		config: conf,
	}
}
func (a *Auth) GenerateJWT(userId uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userId.String(),
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})
	tokenString, err := token.SignedString([]byte(a.config.Jwt.Key))
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}
	return tokenString, err
}
func (a *Auth) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}
	user, err := a.user.GetByEmail(body.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email or password"})
		return
	}
	tokenString, err := a.GenerateJWT(user.UserId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

func (a *Auth) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", "", -1, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func (a *Auth) Signup(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if c.ShouldBindJSON(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}
	if existing, _ := a.user.GetByEmail(body.Email); existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already registered"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to hash password"})
		return
	}
	userId := uuid.New()
	err = a.user.Insert(userId, body.Email, string(hash))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to insert user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "signup successful"})

}
