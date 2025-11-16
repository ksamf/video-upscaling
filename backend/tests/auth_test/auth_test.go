package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/auth"
	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/ksamf/video-upscaling/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuth_Signup_Success(t *testing.T) {
	t.Parallel()
	conf := &config.Config{}
	mockDB := testutil.MockUser{
		InsertFunc:     func(userId uuid.UUID, email, hash string) error { return nil },
		GetByEmailFunc: func(email string) (*database.User, error) { return nil, nil },
	}
	a := auth.New(&mockDB, conf)
	r := gin.New()
	r.POST("/signup", a.Signup)
	signupBody := database.User{
		Email:    "test@example.com",
		Password: "123456",
	}
	b, _ := json.Marshal(signupBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "signup successful")
}

func TestAuth_Signup_EmailExist(t *testing.T) {
	t.Parallel()

	conf := &config.Config{
		Jwt: config.JwtConfig{Key: "secretkey"},
	}
	mockDB := testutil.MockUser{
		InsertFunc: func(userId uuid.UUID, email, hash string) error { return nil },
		GetByEmailFunc: func(email string) (*database.User, error) {
			return &database.User{
				Email:    "test@example.com",
				Password: "123456",
			}, nil
		},
	}
	a := auth.New(&mockDB, conf)
	r := gin.New()
	r.POST("/signup", a.Signup)
	signupBody := database.User{
		Email:    "test@example.com",
		Password: "123456",
	}
	b, _ := json.Marshal(signupBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "email already registered")
}
func TestAuth_Signup_EmptyBody(t *testing.T) {
	t.Parallel()

	conf := &config.Config{
		Jwt: config.JwtConfig{Key: "secretkey"},
	}
	mockDB := testutil.MockUser{
		InsertFunc: func(userId uuid.UUID, email, hash string) error { return nil },
		GetByEmailFunc: func(email string) (*database.User, error) {
			return &database.User{
				Email:    "test@example.com",
				Password: "123456",
			}, nil
		},
	}
	a := auth.New(&mockDB, conf)
	r := gin.New()
	r.POST("/signup", a.Signup)
	signupBody := database.User{}
	b, _ := json.Marshal(signupBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "email already registered")
}
func TestAuth_Login_Success(t *testing.T) {
	t.Parallel()

	conf := &config.Config{
		Jwt: config.JwtConfig{Key: "secretkey"},
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	mockDB := testutil.MockUser{
		GetByEmailFunc: func(email string) (*database.User, error) {
			return &database.User{
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
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "login successful")

	cookies := w.Result().Cookies()
	require.NotEmpty(t, cookies)
	cookie := cookies[0]
	assert.Equal(t, "Authorization", cookie.Name)
	assert.NotEmpty(t, cookie.Value)
}
func TestAuth_Login_WrongPassword(t *testing.T) {
	t.Parallel()

	conf := &config.Config{
		Jwt: config.JwtConfig{Key: "secretkey"},
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	mockDB := testutil.MockUser{
		GetByEmailFunc: func(email string) (*database.User, error) {
			return &database.User{
				Email:    "test@example.com",
				Password: string(hashed),
			}, nil
		},
	}
	a := auth.New(&mockDB, conf)
	r := gin.New()
	r.POST("/login", a.Login)
	signupBody := database.User{
		Email:    "test@example.com",
		Password: "654321",
	}
	b, _ := json.Marshal(signupBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid email or password")

	cookies := w.Result().Cookies()
	require.Empty(t, cookies)
}
func TestAuth_Login_WithoutPassword(t *testing.T) {
	t.Parallel()

	conf := &config.Config{
		Jwt: config.JwtConfig{Key: "secretkey"},
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	mockDB := testutil.MockUser{
		GetByEmailFunc: func(email string) (*database.User, error) {
			return &database.User{
				Email:    "test@example.com",
				Password: string(hashed),
			}, nil
		},
	}
	a := auth.New(&mockDB, conf)
	r := gin.New()
	r.POST("/login", a.Login)
	signupBody := database.User{
		Email:    "test@example.com",
		Password: "",
	}
	b, _ := json.Marshal(signupBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid email or password")

	cookies := w.Result().Cookies()
	require.Empty(t, cookies)
}
func TestAuth_Logout_Success(t *testing.T) {
	t.Parallel()

	conf := &config.Config{
		Jwt: config.JwtConfig{Key: "secretkey"},
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	mockDB := testutil.MockUser{
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
	signupBody := database.User{
		Email:    "test@example.com",
		Password: "123456",
	}
	b, _ := json.Marshal(signupBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	cookie := w.Result().Cookies()[0]

	r.POST("/logout", a.AuthMiddleware, a.Logout)
	w2 := httptest.NewRecorder()
	req4, _ := http.NewRequest("POST", "/logout", nil)
	req4.AddCookie(cookie)
	r.ServeHTTP(w2, req4)
	assert.Equal(t, http.StatusOK, w2.Code)
	logoutCookies := w2.Result().Cookies()
	if len(logoutCookies) > 0 {
		assert.Empty(t, logoutCookies[0].Value)
	}
}
func TestAuth_Middleware_Success(t *testing.T) {
	t.Parallel()
	conf := &config.Config{
		Jwt: config.JwtConfig{Key: "secretkey"},
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	mockDB := testutil.MockUser{
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
	cookie := cookies[0]

	r.GET("/protected", a.AuthMiddleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/protected", nil)
	req2.AddCookie(cookie)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
func TestAuth_Middleware_NotLogin(t *testing.T) {
	t.Parallel()
	conf := &config.Config{
		Jwt: config.JwtConfig{Key: "secretkey"},
	}
	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	mockDB := testutil.MockUser{
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
	r.GET("/protected", a.AuthMiddleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)
	assert.Empty(t, w.Result().Cookies())
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
