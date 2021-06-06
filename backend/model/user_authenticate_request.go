package model

type UserAuthenticateRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}
