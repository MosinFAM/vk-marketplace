package handlers

import (
	"net/http"
	"strconv"

	"github.com/MosinFAM/vk-marketplace/internal/auth"
	"github.com/MosinFAM/vk-marketplace/internal/models"
	"github.com/MosinFAM/vk-marketplace/internal/repository"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Repo repository.Repository
}

func validateCredentials(username, password string) bool {
	return len(username) >= 3 && len(username) <= 30 && len(password) >= 6 && len(password) <= 100
}

func validateAd(ad models.Ad) bool {
	return len(ad.Title) > 0 && len(ad.Title) <= 100 &&
		len(ad.Description) <= 1000 &&
		len(ad.ImageURL) <= 300 &&
		ad.Price > 0
}

// ======= AUTH =======

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.AuthRequest true "username and password"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /register [post]
func (h *Handler) Register(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil || !validateCredentials(req.Username, req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid username or password format"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	id, err := h.Repo.RegisterUser(req.Username, string(hash))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	token, _ := auth.GenerateToken(id)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       id,
			"username": req.Username,
		},
	})
}

// Login godoc
// @Summary Login user
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.AuthRequest true "username and password"
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req models.AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := h.Repo.GetUserByUsername(req.Username)
	if err != nil || user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect credentials"})
		return
	}

	token, _ := auth.GenerateToken(user.ID)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

// ======= ADS =======

// CreateAd godoc
// @Summary Create a new ad
// @Tags ads
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param input body models.Ad true "ad object"
// @Success 200 {object} models.Ad
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /ads [post]
func (h *Handler) CreateAd(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var ad models.Ad
	if err := c.ShouldBindJSON(&ad); err != nil || !validateAd(ad) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ad"})
		return
	}
	ad.UserID = userID

	created, err := h.Repo.CreateAd(ad)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ad"})
		return
	}
	c.JSON(http.StatusOK, created)
}

// ListAds godoc
// @Summary Get list of ads with filters & pagination
// @Tags ads
// @Security ApiKeyAuth
// @Produce json
// @Param sort_by query string false "Sort by field: created_at or price"
// @Param sort_order query string false "asc or desc"
// @Param min_price query number false "Min price"
// @Param max_price query number false "Max price"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} models.Ad
// @Failure 500 {object} models.ErrorResponse
// @Router /ads [get]
func (h *Handler) ListAds(c *gin.Context) {
	var f repository.AdsFilter

	if s := c.Query("sort_by"); s != "" {
		f.SortBy = s
	}
	if s := c.Query("sort_order"); s != "" {
		f.SortOrder = s
	}
	if min := c.Query("min_price"); min != "" {
		if val, err := strconv.ParseFloat(min, 64); err == nil {
			f.MinPrice = &val
		}
	}
	if max := c.Query("max_price"); max != "" {
		if val, err := strconv.ParseFloat(max, 64); err == nil {
			f.MaxPrice = &val
		}
	}
	if limit := c.DefaultQuery("limit", "10"); limit != "" {
		f.Limit, _ = strconv.Atoi(limit)
	}
	if offset := c.DefaultQuery("offset", "0"); offset != "" {
		f.Offset, _ = strconv.Atoi(offset)
	}

	if uid, ok := c.Get("user_id"); ok {
		id := uid.(string)
		f.UserID = &id
	}

	ads, err := h.Repo.ListAds(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	c.JSON(http.StatusOK, ads)
}

// GetAdByID godoc
// @Summary Get a single ad by ID
// @Tags ads
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Ad ID"
// @Success 200 {object} models.Ad
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /ads/{id} [get]
func (h *Handler) GetAdByID(c *gin.Context) {
	id := c.Param("id")
	ad, err := h.Repo.GetAdByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if ad == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ad not found"})
		return
	}

	userID := c.GetString("user_id")
	if userID != "" && userID == ad.UserID {
		ad.IsOwner = true
	}

	c.JSON(http.StatusOK, ad)
}
