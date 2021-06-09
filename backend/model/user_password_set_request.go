package model

type UserPasswordSetRequest struct {
	Token    *string `json:"token"`
	Password *string `json:"password"`
}
