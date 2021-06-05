package service

import (
	"crypto/sha256"
	"fmt"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/utils"
	"github.com/syncloud/redirect/validator"
	"time"
)

const (
	PremiumStatusInactive = 1
	PremiumStatusPending  = 2
	PremiumStatusActive   = 3
)

type UsersDb interface {
	GetUserByEmail(email string) (*model.User, error)
	GetUserByUpdateToken(updateToken string) (*model.User, error)
	InsertUser(user *model.User) (int64, error)
	GetUser(id int64) (*model.User, error)
	UpdateUser(user *model.User) error
	DeleteUser(userId int64) error
}

type UsersActions interface {
	GetActivateAction(token string) (*model.Action, error)
	UpsertActivateAction(userId int64) (*model.Action, error)
	DeleteActions(userId int64) error
}

type UsersMail interface {
	SendActivate(to string, token string) error
	SendPremiumRequest(to string) error
}
type Users struct {
	db              UsersDb
	activateByEmail bool
	actions         UsersActions
	usersMail       UsersMail
}

func NewUsers(db UsersDb, activateByEmail bool, actions UsersActions, usersMail *Mail) *Users {
	return &Users{db: db, activateByEmail: activateByEmail, actions: actions, usersMail: usersMail}
}

func (u *Users) Authenticate(email *string, password *string) (*model.User, error) {
	fieldValidator := validator.New()

	emailLower := fieldValidator.Email(email)
	passwordChecked := fieldValidator.Password(password)
	if fieldValidator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: fieldValidator.ToParametersMessages()}
	}

	user, err := u.db.GetUserByEmail(*emailLower)
	if err != nil || user == nil || !user.Active || hash(*passwordChecked) != user.PasswordHash {
		return nil, &model.ServiceError{InternalError: fmt.Errorf("authentication failed")}
	}

	return user, nil
}

func (u *Users) Activate(token string) error {

	action, err := u.actions.GetActivateAction(token)
	if err != nil {
		return err
	}

	user, err := u.db.GetUser(action.UserId)
	if err != nil {
		return err
	}
	if user == nil {
		return &model.ServiceError{InternalError: fmt.Errorf("invalid activation token")}
	}

	if user.Active {
		return &model.ServiceError{InternalError: fmt.Errorf("user is active already")}
	}

	user.Active = true
	err = u.db.UpdateUser(user)
	return err
}

func (u *Users) Save(user *model.User) error {
	return u.db.UpdateUser(user)
}

func (u *Users) GetUserByEmail(userEmail string) (*model.User, error) {
	return u.db.GetUserByEmail(userEmail)
}

func (u *Users) GetUserByUpdateToken(updateToken string) (*model.User, error) {
	return u.db.GetUserByUpdateToken(updateToken)
}

func (u *Users) Delete(userId int64) error {
	err := u.actions.DeleteActions(userId)
	if err != nil {
		return err
	}
	return u.db.DeleteUser(userId)
}

func (u *Users) CreateNewUser(request model.UserCreateRequest) (*model.User, error) {
	fieldValidator := validator.New()
	email := fieldValidator.Email(request.Email)
	password := fieldValidator.NewPassword(request.Password)
	if fieldValidator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: fieldValidator.ToParametersMessages()}
	}
	userByEmail, err := u.db.GetUserByEmail(*email)
	if err != nil {
		return nil, err
	}
	if userByEmail != nil {
		return nil, model.SingleParameterError("email", "Email is already registered")
	}

	updateToken := utils.Uuid()
	user := &model.User{Email: *email, PasswordHash: hash(*password), Active: !u.activateByEmail, UpdateToken: updateToken, PremiumStatusId: PremiumStatusInactive, Timestamp: time.Now()}

	userId, err := u.db.InsertUser(user)
	if err != nil {
		return nil, err
	}

	if u.activateByEmail {
		action, err := u.actions.UpsertActivateAction(userId)
		if err != nil {
			return nil, err
		}
		err = u.usersMail.SendActivate(user.Email, action.Token)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (u *Users) RequestPremiumAccount(user *model.User) error {
	if user.PremiumStatusId != PremiumStatusInactive {
		return fmt.Errorf("premium account is already requested")
	}
	user.PremiumStatusId = PremiumStatusPending
	err := u.db.UpdateUser(user)
	if err != nil {
		return err
	}
	return u.usersMail.SendPremiumRequest(user.Email)
}

func hash(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}
