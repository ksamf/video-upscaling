package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/ksamf/video-upscaling/backend/internal/handler"
	"github.com/ksamf/video-upscaling/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestUserHandler_Get_Success(t *testing.T) {
	t.Parallel()
	id := uuid.New()
	mockUser := &testutil.MockUser{
		GetByIdFunc: func(uid uuid.UUID) (*database.User, error) {
			return &database.User{
				UserId: id,
				Email:  "test@example.com",
			}, nil
		},
	}
	mockS3 := &testutil.MockStorage{
		GetURLFunc: func(id uuid.UUID) string { return "https://example.com/" + id.String() },
	}

	h := handler.NewUserHandlers(mockUser, mockS3)

	router := gin.New()
	router.GET("/user/:id", h.Get)

	req, _ := http.NewRequest(http.MethodGet, "/user/"+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp database.User
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "test@example.com", resp.Email)
}
func TestUserHandler_Get_InvalidUUID(t *testing.T) {
	t.Parallel()

	mockUser := &testutil.MockUser{}
	mockStorage := &testutil.MockStorage{}
	h := handler.NewUserHandlers(mockUser, mockStorage)

	router := gin.New()
	router.GET("/user/:id", h.Get)

	req, _ := http.NewRequest(http.MethodGet, "/user/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
func TestUserHandler_Delete_Success(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	mockUser := &testutil.MockUser{
		DeleteFunc: func(uid uuid.UUID) error {
			return nil
		},
	}
	mockS3 := &testutil.MockStorage{
		GetURLFunc: func(id uuid.UUID) string { return "https://example.com/" + id.String() },
	}

	h := handler.NewUserHandlers(mockUser, mockS3)

	router := gin.New()
	router.DELETE("/user/:id", h.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/user/"+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
func TestUserHandler_Delete_InvalidUUID(t *testing.T) {
	t.Parallel()

	mockUser := &testutil.MockUser{}
	mockStorage := &testutil.MockStorage{}
	h := handler.NewUserHandlers(mockUser, mockStorage)

	router := gin.New()
	router.DELETE("/user/:id", h.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/user/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUserHandler_Get_Recommendation(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	mockUser := &testutil.MockUser{
		GetMixedRecommendationFunc: func(uid uuid.UUID, getURL func(uuid.UUID) string) ([]*database.Video, error) {
			return []*database.Video{
				{
					VideoId: uuid.New(),
					Name:    "Recommended video",
				},
			}, nil
		},
	}
	mockS3 := &testutil.MockStorage{
		GetURLFunc: func(id uuid.UUID) string { return "https://example.com/" + id.String() },
	}

	h := handler.NewUserHandlers(mockUser, mockS3)

	router := gin.New()
	router.GET("/user/:id/recommendation", h.Recommendation)

	req, _ := http.NewRequest(http.MethodGet, "/user/"+id.String()+"/recommendation", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []*database.Video
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "Recommended video", resp[0].Name)
}

func TestUserHandler_Get_Videos(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	mockUser := &testutil.MockUser{
		VideosFunc: func(id uuid.UUID, getURL func(uuid.UUID) string) ([]*database.Video, error) {
			return []*database.Video{
				{
					VideoId: uuid.New(),
					UserId:  id,
					Name:    "User video",
				},
			}, nil
		},
	}

	mockS3 := &testutil.MockStorage{
		GetURLFunc: func(id uuid.UUID) string { return "https://example.com/" + id.String() },
	}

	h := handler.NewUserHandlers(mockUser, mockS3)

	router := gin.New()
	router.GET("/user/:id/video", h.Videos)

	req, _ := http.NewRequest(http.MethodGet, "/user/"+id.String()+"/video", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []*database.Video
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	for _, r := range resp {
		assert.Equal(t, id, r.UserId)
	}
}
