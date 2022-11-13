package main

import "github.com/syncloud/redirect/model"

type TestUsers struct {
}

func (u *TestUsers) CreateNewUser(request model.UserCreateRequest) (*model.User, error) {
	return &model.User{}, nil
}

func (u *TestUsers) Authenticate(email *string, password *string) (*model.User, error) {
	return &model.User{}, nil
}

func (u *TestUsers) GetUserByUpdateToken(updateToken string) (*model.User, error) {
	return &model.User{}, nil
}
