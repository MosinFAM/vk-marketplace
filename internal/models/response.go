package models

type ErrorResponse struct {
	Error string `json:"error" example:"Invalid credentials"`
}

type UserResponse struct {
	Token string `json:"token" example:"your.jwt.token"`
	User  struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}
