package service

import (
	"crypto/sha256"
	"fmt"
	"github.com/syncloud/redirect/db"
	"github.com/syncloud/redirect/model"
)

type Users struct {
	db *db.MySql
}

func NewUsers(db *db.MySql) *Users {
	return &Users{db: db}
}

func (u *Users) Authenticate(email *string, password *string) (*model.User, error) {
	validator := NewValidator()

	emailLower := validator.email(email)
	passwordChecked := validator.password(password)
	if validator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: validator.ToParametersMessages()}
	}

	user, err := u.db.GetUserByEmail(*emailLower)
	if err != nil || user == nil || !user.Active || hash(*passwordChecked) != user.PasswordHash {
		return nil, &model.ServiceError{InternalError: fmt.Errorf("authentication failed")}
	}

	return user, nil
}

func hash(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}
