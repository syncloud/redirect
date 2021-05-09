package model

type UserGetRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}
