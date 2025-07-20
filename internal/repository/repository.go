package repository

import "github.com/MosinFAM/vk-marketplace/internal/models"

type AdsFilter struct {
	SortBy    string
	SortOrder string
	MinPrice  *float64
	MaxPrice  *float64
	Limit     int
	Offset    int
	UserID    *string
}

// go install go.uber.org/mock/mockgen@latest
//
//go:generate mockgen -source=repository.go -destination=repo_mock.go -package=repository Repository
type Repository interface {
	CreateAd(ad models.Ad) (models.Ad, error)
	ListAds(filter AdsFilter) ([]models.Ad, error)
	GetAdByID(id string) (*models.Ad, error)
	RegisterUser(username, passwordHash string) (string, error)
	GetUserByUsername(username string) (*models.User, error)
}
