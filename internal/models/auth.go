package models

type AuthRequest struct {
	Username string `json:"username" example:"john_doe"`
	Password string `json:"password" example:"secure123"`
}
