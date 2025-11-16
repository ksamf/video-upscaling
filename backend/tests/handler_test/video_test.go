package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/broker"
	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/ksamf/video-upscaling/backend/internal/handler"
	"github.com/ksamf/video-upscaling/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestVideoHandler_Get_All_Success(t *testing.T) {
	t.Parallel()

	mockVideo := &testutil.MockVideo{
		GetAllFunc: func(limit, offset string, getURL func(uuid.UUID) string) ([]*database.Video, error) {
			return []*database.Video{
				{VideoId: uuid.New()},
				{VideoId: uuid.New()},
				{VideoId: uuid.New()},
				{VideoId: uuid.New()},
			}, nil
		},
	}
	mockAction := &testutil.MockAction{}
	mockS3 := &testutil.MockStorage{
		GetURLFunc: func(id uuid.UUID) string { return "https://example.com/" + id.String() },
	}
	mockCache := &testutil.MockCache{
		GetFunc: func(key string) string { return "" },
		SetFunc: func(key string, value any, exp time.Duration) {},
	}
	conf := &config.Config{}
	mockBroker := &testutil.MockBroker{}

	h := handler.NewVideoHandlers(mockVideo, mockAction, conf, mockS3, mockCache, mockBroker)

	router := gin.New()
	router.GET("/video", h.GetAll)

	req, _ := http.NewRequest(http.MethodGet, "/video", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []database.Video
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Len(t, resp, 4)
}
func TestVideoHandler_Get_Success(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	mockVideo := &testutil.MockVideo{
		GetByIDFunc: func(id uuid.UUID, getURL func(uuid.UUID) string) (*database.Video, error) {
			return &database.Video{
				VideoId: id,
			}, nil
		},
	}
	mockAction := &testutil.MockAction{
		InsertFunc: func(userId, videoId uuid.UUID) error {
			return nil
		},
	}

	mockS3 := &testutil.MockStorage{}
	mockCache := &testutil.MockCache{
		GetFunc:  func(key string) string { return "" },
		SetFunc:  func(key string, value any, exp time.Duration) {},
		IncrFunc: func(key string) {},
	}
	conf := &config.Config{}
	mockBroker := &testutil.MockBroker{}

	h := handler.NewVideoHandlers(mockVideo, mockAction, conf, mockS3, mockCache, mockBroker)

	router := gin.New()
	router.GET("/video/:id", h.Get)

	req, _ := http.NewRequest(http.MethodGet, "/video/"+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp database.Video
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, id, resp.VideoId)
}
func TestVideoHandler_Get_InvalidUUID(t *testing.T) {
	t.Parallel()

	mockVideo := &testutil.MockVideo{}

	mockS3 := &testutil.MockStorage{}
	mockCache := &testutil.MockCache{}
	conf := &config.Config{}
	mockBroker := &testutil.MockBroker{}
	mockAction := &testutil.MockAction{}

	h := handler.NewVideoHandlers(mockVideo, mockAction, conf, mockS3, mockCache, mockBroker)

	router := gin.New()
	router.GET("/video/:id", h.Get)

	req, _ := http.NewRequest(http.MethodGet, "/video/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
func TestVideoHandler_Delete_Success(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	mockVideo := &testutil.MockVideo{
		DeleteFunc: func(id uuid.UUID) error { return nil },
	}
	mockS3 := &testutil.MockStorage{
		DeleteObjectFunc: func(object string) error { return nil },
	}
	mockCache := &testutil.MockCache{}
	conf := &config.Config{}
	mockBroker := &testutil.MockBroker{}
	mockAction := &testutil.MockAction{}

	h := handler.NewVideoHandlers(mockVideo, mockAction, conf, mockS3, mockCache, mockBroker)

	router := gin.New()
	router.DELETE("/video/:id", h.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/video/"+id.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
func TestVideoHandler_Delete_InvalidUUID(t *testing.T) {
	t.Parallel()

	mockVideo := &testutil.MockVideo{}
	mockS3 := &testutil.MockStorage{}
	mockCache := &testutil.MockCache{}
	conf := &config.Config{}
	mockBroker := &testutil.MockBroker{}
	mockAction := &testutil.MockAction{}
	h := handler.NewVideoHandlers(mockVideo, mockAction, conf, mockS3, mockCache, mockBroker)

	router := gin.New()
	router.DELETE("/video/:id", h.Delete)

	req, _ := http.NewRequest(http.MethodDelete, "/video/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestVideoHandler_Upload_Success(t *testing.T) {
	conf := &config.Config{}

	mockDB := testutil.MockVideo{
		InsertFunc: func(video *database.Video) error {
			return nil
		},
	}
	calledS3 := false
	mockS3 := &testutil.MockStorage{
		PutObjectFunc: func(key string, r io.Reader) error {
			calledS3 = true
			return nil
		},
	}
	mockBroker := &testutil.MockBroker{
		WriteFunc: func(msg broker.KafkaMessage) error {
			return nil
		},
	}
	tmpfile, err := os.CreateTemp("", "testvideo*.mp4")
	assert.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	io.WriteString(tmpfile, "fake video data")
	tmpfile.Close()
	r := gin.New()

	body2 := &bytes.Buffer{}
	writer := multipart.NewWriter(body2)
	part, _ := writer.CreateFormFile("file", "example.mp4")

	file, _ := os.Open(tmpfile.Name())
	io.Copy(part, file)
	file.Close()
	writer.Close()
	auth, cookie := testutil.Login()
	mockAction := &testutil.MockAction{}
	h := handler.NewVideoHandlers(&mockDB, mockAction, conf, mockS3, nil, mockBroker)

	r.POST("/video", auth.AuthMiddleware, h.Upload)

	req, _ := http.NewRequest(http.MethodPost, "/video", body2)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	fmt.Println(w.Body.String())
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "upload")
	assert.True(t, calledS3, "S3.PutObject должен быть вызван")
}
func TestVideoHandler_Upload_NoFile(t *testing.T) {
	mockVideo := &testutil.MockVideo{}
	mockS3 := &testutil.MockStorage{}
	mockCache := &testutil.MockCache{}
	conf := &config.Config{}
	mockBroker := &testutil.MockBroker{}
	mockAction := &testutil.MockAction{}
	h := handler.NewVideoHandlers(mockVideo, mockAction, conf, mockS3, mockCache, mockBroker)

	router := gin.New()
	router.POST("/upload", h.Upload)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
func TestVideoHandler_Upload_InvalidFormat(t *testing.T) {
	r := gin.New()
	auth, cookie := testutil.Login()

	body2 := &bytes.Buffer{}
	writer := multipart.NewWriter(body2)
	part, _ := writer.CreateFormFile("file", "example.txt")
	part.Write([]byte("fake data"))
	writer.Close()
	mockVideo := &testutil.MockVideo{}
	mockS3 := &testutil.MockStorage{}
	mockCache := &testutil.MockCache{}
	conf := &config.Config{}
	mockBroker := &testutil.MockBroker{}
	mockAction := &testutil.MockAction{}

	h := handler.NewVideoHandlers(mockVideo, mockAction, conf, mockS3, mockCache, mockBroker)

	r.POST("/video", auth.AuthMiddleware, h.Upload)

	req2, _ := http.NewRequest(http.MethodPost, "/video", body2)
	req2.Header.Set("Content-Type", writer.FormDataContentType())
	req2.AddCookie(cookie)

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
	assert.Contains(t, w2.Body.String(), "failed format video")
}
