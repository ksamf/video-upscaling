package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/auth"
	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"golang.org/x/crypto/bcrypt"
)

func Login() (*auth.Auth, *http.Cookie) {
	conf := &config.Config{
		Jwt: config.JwtConfig{Key: "secretkey"},
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	mockDB := MockUser{
		InsertFunc: func(userId uuid.UUID, email, password string) error {
			return nil
		},
		GetByEmailFunc: func(email string) (*database.User, error) {
			return &database.User{
				Email:    "test@example.com",
				Password: string(hashed),
			}, nil
		},
		GetByIdFunc: func(id uuid.UUID) (*database.User, error) {
			return &database.User{
				UserId:   id,
				Email:    "test@example.com",
				Password: string(hashed),
			}, nil
		},
	}

	a := auth.New(&mockDB, conf)
	r := gin.New()
	r.POST("/login", a.Login)
	body := database.User{
		Email:    "test@example.com",
		Password: "123456",
	}
	b, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(b))
	r.ServeHTTP(w, req)
	req.Header.Set("Content-Type", "application/json")
	cookies := w.Result().Cookies()
	return a, cookies[0]
}
