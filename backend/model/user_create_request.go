package model

type UserCreateRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}
