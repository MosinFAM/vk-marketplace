package models

type AuthRequest struct {
	Username string `json:"username" example:"sasha_fil"`
	Password string `json:"password" example:"secure123"`
}
