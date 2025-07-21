package models

import "time"

type Ad struct {
	ID          string    `json:"id" example:"adcacsac123"`
	Title       string    `json:"title" example:"Buy GTA6"`
	Description string    `json:"description" example:"Buy GTA6"`
	ImageURL    string    `json:"image_url" example:"http://example.com/image.jpg"`
	Price       float64   `json:"price" example:"888"`
	CreatedAt   time.Time `json:"created_at" example:"2025-07-19T12:34:56Z"`
	UserID      string    `json:"user_id" example:"user123"`
	Username    string    `json:"username,omitempty" example:"sasha_fil"`
	IsOwner     bool      `json:"is_owner,omitempty" example:"true"`
}
