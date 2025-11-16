package interfaces

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Auth interface {
	GenerateJWT(userId uuid.UUID) (string, error)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	Signup(c *gin.Context)
	AuthMiddleware(c *gin.Context)
	OptionalAuthMiddleware(c *gin.Context)
}
