package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MosinFAM/vk-marketplace/internal/models"
	"github.com/MosinFAM/vk-marketplace/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

// helper для создания gin.Context с request и response recorder
func getTestContext(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	return c, w
}

// helper для создания gin.Context c user_id
func getTestContextWithUser(method, path string, body []byte, userID string) (*gin.Context, *httptest.ResponseRecorder) {
	c, w := getTestContext(method, path, body)
	if userID != "" {
		c.Set("user_id", userID)
	}
	return c, w
}

func TestHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockRepository(ctrl)
	h := &Handler{Repo: mockRepo}

	tests := []struct {
		name       string
		reqBody    interface{}
		setupMock  func()
		wantStatus int
	}{
		{
			name:    "success",
			reqBody: models.AuthRequest{Username: "testuser", Password: "password123"},
			setupMock: func() {
				mockRepo.EXPECT().RegisterUser("testuser", gomock.Any()).Return("user-id-123", nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad request invalid json",
			reqBody:    nil, // special case: invalid json
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bad request invalid credentials",
			reqBody:    models.AuthRequest{Username: "a", Password: "123"},
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "internal error on repo",
			reqBody: models.AuthRequest{Username: "testuser", Password: "password123"},
			setupMock: func() {
				mockRepo.EXPECT().RegisterUser("testuser", gomock.Any()).Return("", errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.reqBody != nil {
				body, _ = json.Marshal(tt.reqBody)
			} else {
				body = []byte(`{invalid json`)
			}

			tt.setupMock()
			c, w := getTestContext("POST", "/register", body)
			h.Register(c)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "user-id-123", resp["user"].(map[string]interface{})["id"])
				assert.NotEmpty(t, resp["token"])
			}
		})
	}
}

func TestHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockRepository(ctrl)
	h := &Handler{Repo: mockRepo}

	tests := []struct {
		name       string
		reqBody    interface{}
		setupMock  func()
		wantStatus int
	}{
		{
			name:    "success",
			reqBody: models.AuthRequest{Username: "testuser", Password: "password123"},
			setupMock: func() {
				hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &models.User{ID: "user-id-123", Username: "testuser", PasswordHash: string(hash)}
				mockRepo.EXPECT().GetUserByUsername("testuser").Return(user, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "bad request invalid json",
			reqBody:    nil,
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "unauthorized wrong password",
			reqBody: models.AuthRequest{Username: "testuser", Password: "wrongpassword"},
			setupMock: func() {
				hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				user := &models.User{ID: "user-id-123", Username: "testuser", PasswordHash: string(hash)}
				mockRepo.EXPECT().GetUserByUsername("testuser").Return(user, nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "unauthorized user not found",
			reqBody: models.AuthRequest{Username: "nouser", Password: "password123"},
			setupMock: func() {
				mockRepo.EXPECT().GetUserByUsername("nouser").Return(nil, nil)
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.reqBody != nil {
				body, _ = json.Marshal(tt.reqBody)
			} else {
				body = []byte(`{invalid json`)
			}

			tt.setupMock()
			c, w := getTestContext("POST", "/login", body)
			h.Login(c)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "user-id-123", resp["user"].(map[string]interface{})["id"])
				assert.NotEmpty(t, resp["token"])
			}
		})
	}
}

func TestHandler_CreateAd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockRepository(ctrl)
	h := &Handler{Repo: mockRepo}

	validAd := models.Ad{
		Title:       "Test Ad",
		Description: "Description",
		ImageURL:    "http://image.url",
		Price:       100,
	}

	tests := []struct {
		name       string
		ad         models.Ad
		userID     string
		setupMock  func()
		wantStatus int
	}{
		{
			name:   "success",
			ad:     validAd,
			userID: "user-id-123",
			setupMock: func() {
				returnedAd := validAd
				returnedAd.ID = "ad-id-123"
				returnedAd.UserID = "user-id-123"
				mockRepo.EXPECT().CreateAd(gomock.AssignableToTypeOf(models.Ad{})).Return(returnedAd, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "unauthorized no user id",
			ad:         validAd,
			userID:     "",
			setupMock:  func() {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad request invalid ad",
			ad:         models.Ad{Title: "", Price: -1},
			userID:     "user-id-123",
			setupMock:  func() {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "internal error on repo",
			ad:     validAd,
			userID: "user-id-123",
			setupMock: func() {
				mockRepo.EXPECT().CreateAd(gomock.AssignableToTypeOf(models.Ad{})).Return(models.Ad{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.ad)
			tt.setupMock()
			c, w := getTestContextWithUser("POST", "/ads", body, tt.userID)
			h.CreateAd(c)
			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantStatus == http.StatusOK {
				var resp models.Ad
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, "ad-id-123", resp.ID)
				assert.Equal(t, "user-id-123", resp.UserID)
			}
		})
	}
}

func TestHandler_GetAdByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockRepository(ctrl)
	h := &Handler{Repo: mockRepo}

	tests := []struct {
		name       string
		ad         *models.Ad
		userID     string
		paramID    string
		mockReturn func()
		wantStatus int
		wantOwner  bool
	}{
		{
			name: "success with owner",
			ad: &models.Ad{
				ID:     "ad-id-123",
				UserID: "user-id-123",
				Title:  "Ad title",
			},
			userID:  "user-id-123",
			paramID: "ad-id-123",
			mockReturn: func() {
				mockRepo.EXPECT().GetAdByID("ad-id-123").Return(&models.Ad{
					ID:     "ad-id-123",
					UserID: "user-id-123",
					Title:  "Ad title",
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantOwner:  true,
		},
		{
			name: "success without owner",
			ad: &models.Ad{
				ID:     "ad-id-123",
				UserID: "other-user",
				Title:  "Ad title",
			},
			userID:  "user-id-999",
			paramID: "ad-id-123",
			mockReturn: func() {
				mockRepo.EXPECT().GetAdByID("ad-id-123").Return(&models.Ad{
					ID:     "ad-id-123",
					UserID: "other-user",
					Title:  "Ad title",
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantOwner:  false,
		},
		{
			name:       "not found",
			userID:     "",
			paramID:    "missing-id",
			mockReturn: func() { mockRepo.EXPECT().GetAdByID("missing-id").Return(nil, nil) },
			wantStatus: http.StatusNotFound,
			wantOwner:  false,
		},
		{
			name:       "db error",
			userID:     "",
			paramID:    "some-id",
			mockReturn: func() { mockRepo.EXPECT().GetAdByID("some-id").Return(nil, errors.New("db error")) },
			wantStatus: http.StatusInternalServerError,
			wantOwner:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockReturn()
			c, w := getTestContext("GET", "/ads/"+tt.paramID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.paramID}}
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}
			h.GetAdByID(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var resp models.Ad
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.paramID, resp.ID)
				assert.Equal(t, tt.wantOwner, resp.IsOwner)
			}
		})
	}
}

func TestHandler_ListAds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepo := repository.NewMockRepository(ctrl)
	h := &Handler{Repo: mockRepo}

	minPrice := 10.5
	maxPrice := 200.0
	limit := 15
	offset := 7
	userID := "user-abc"

	tests := []struct {
		name       string
		query      string
		userID     string
		expected   repository.AdsFilter
		mockReturn func()
		wantStatus int
		wantLen    int
	}{
		{
			name:   "success with all query params and user_id",
			query:  "sort_by=price&sort_order=asc&min_price=10.5&max_price=200&limit=15&offset=7",
			userID: userID,
			expected: repository.AdsFilter{
				SortBy:    "price",
				SortOrder: "asc",
				MinPrice:  &minPrice,
				MaxPrice:  &maxPrice,
				Limit:     limit,
				Offset:    offset,
				UserID:    &userID,
			},
			mockReturn: func() {
				mockRepo.EXPECT().ListAds(repository.AdsFilter{
					SortBy:    "price",
					SortOrder: "asc",
					MinPrice:  &minPrice,
					MaxPrice:  &maxPrice,
					Limit:     limit,
					Offset:    offset,
					UserID:    &userID,
				}).Return([]models.Ad{
					{ID: "ad1", Title: "Ad One", Price: 50},
					{ID: "ad2", Title: "Ad Two", Price: 150},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name:   "success with default limit and offset, no user_id",
			query:  "",
			userID: "",
			expected: repository.AdsFilter{
				Limit:  10,
				Offset: 0,
			},
			mockReturn: func() {
				mockRepo.EXPECT().ListAds(repository.AdsFilter{
					Limit:  10,
					Offset: 0,
				}).Return([]models.Ad{}, nil)
			},
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name:   "ignores invalid min_price and max_price",
			query:  "min_price=abc&max_price=def",
			userID: "",
			expected: repository.AdsFilter{
				Limit:  10,
				Offset: 0,
			},
			mockReturn: func() {
				mockRepo.EXPECT().ListAds(repository.AdsFilter{
					Limit:  10,
					Offset: 0,
				}).Return([]models.Ad{}, nil)
			},
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name:   "repo returns error",
			query:  "",
			userID: "",
			expected: repository.AdsFilter{
				Limit:  10,
				Offset: 0,
			},
			mockReturn: func() {
				mockRepo.EXPECT().ListAds(repository.AdsFilter{
					Limit:  10,
					Offset: 0,
				}).Return(nil, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockReturn()
			c, w := getTestContext("GET", "/ads?"+tt.query, nil)
			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}
			h.ListAds(c)
			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantStatus == http.StatusOK {
				var ads []models.Ad
				err := json.Unmarshal(w.Body.Bytes(), &ads)
				assert.NoError(t, err)
				assert.Len(t, ads, tt.wantLen)
			}
		})
	}
}
