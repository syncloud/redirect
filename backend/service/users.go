package service

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/utils"
	"github.com/syncloud/redirect/validation"
	"log"
	"time"
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
	GetPasswordAction(token string) (*model.Action, error)
	DeleteAction(actionId uint64) error
	UpsertPasswordAction(userId int64) (*model.Action, error)
}

type UsersMail interface {
	SendActivate(to string, token string) error
	SendPlanSubscribed(to string) error
	SendPlanUnSubscribed(to string) error
	SendSetPassword(to string) error
	SendResetPassword(to string, token string) error
}

type Subscriptions interface {
	Unsubscribe(id string) error
}

type Users struct {
	db              UsersDb
	activateByEmail bool
	actions         UsersActions
	usersMail       UsersMail
	subscriptions   Subscriptions
}

func NewUsers(db UsersDb,
	activateByEmail bool,
	actions UsersActions,
	usersMail *Mail,
	subscriptions Subscriptions,
) *Users {
	return &Users{
		db:              db,
		activateByEmail: activateByEmail,
		actions:         actions,
		usersMail:       usersMail,
		subscriptions:   subscriptions,
	}
}

func (u *Users) Authenticate(email *string, password *string) (*model.User, error) {
	fieldValidator := validation.New()
	emailLower := fieldValidator.Email(email)
	passwordChecked := fieldValidator.Password(password)
	if fieldValidator.HasErrors() {
		return nil, &model.ParameterError{ParameterErrors: fieldValidator.ToParametersMessages()}
	}

	user, err := u.db.GetUserByEmail(*emailLower)
	if err != nil || user == nil || !user.Active || Hash(*passwordChecked) != user.PasswordHash {
		return nil, model.NewServiceError("authentication failed")
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
		return model.NewServiceError("invalid activation token")
	}

	if user.Active {
		return model.NewServiceError("user is already active")
	}

	user.Active = true
	err = u.db.UpdateUser(user)
	if err != nil {
		return err
	}
	err = u.actions.DeleteAction(action.Id)
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
	fieldValidator := validation.New()
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
	user := &model.User{
		Email:               *email,
		PasswordHash:        Hash(*password),
		Active:              !u.activateByEmail,
		UpdateToken:         updateToken,
		Timestamp:           time.Now(),
		NotificationEnabled: true,
	}

	userId, err := u.db.InsertUser(user)
	if err != nil {
		return nil, err
	}
	log.Printf("user created")

	if u.activateByEmail {
		action, err := u.actions.UpsertActivateAction(userId)
		if err != nil {
			return nil, err
		}
		err = u.usersMail.SendActivate(user.Email, action.Token)
		if err != nil {
			return nil, err
		}
		log.Printf("activation email sent")
	}

	return user, nil
}

func (u *Users) Unsubscribe(user *model.User) error {
	if !user.IsSubscribed() {
		return fmt.Errorf("you have no existing subscrition, please contact support")
	}
	if user.IsPayPal() {
		err := u.subscriptions.Unsubscribe(*user.SubscriptionId)
		if err != nil {
			return err
		}
	}
	user.UnSubscribe(time.Now())
	err := u.db.UpdateUser(user)
	if err != nil {
		return err
	}
	return u.usersMail.SendPlanUnSubscribed(user.Email)
}

func (u *Users) Subscribe(user *model.User, subscriptionId string, subscriptionType int) error {
	if user.IsSubscribed() {
		return fmt.Errorf("you have an existing subscrition, please contact support")
	}
	user.Subscribe(subscriptionId, subscriptionType)
	err := u.db.UpdateUser(user)
	if err != nil {
		return err
	}
	return u.usersMail.SendPlanSubscribed(user.Email)
}

func (u *Users) RequestPasswordReset(email string) (*string, error) {
	user, err := u.GetUserByEmail(email)
	if err != nil {
		log.Println("unable to get a user", err)
		return nil, errors.New("invalid request")
	}

	if user != nil && user.Active {
		action, err := u.actions.UpsertPasswordAction(user.Id)
		if err != nil {
			log.Println("unable to upsert action", err)
			return nil, errors.New("invalid request")
		}
		err = u.usersMail.SendResetPassword(user.Email, action.Token)
		if err != nil {
			log.Println("unable to send mail", err)
			return nil, errors.New("invalid request")
		}
		return &action.Token, nil
	}
	return nil, nil
}

func (u *Users) UserSetPassword(request *model.UserPasswordSetRequest) error {
	fieldValidator := validation.New()
	fieldValidator.Token(request.Token)
	password := fieldValidator.NewPassword(request.Password)
	if fieldValidator.HasErrors() {
		return &model.ParameterError{ParameterErrors: fieldValidator.ToParametersMessages()}
	}

	action, err := u.actions.GetPasswordAction(*request.Token)
	if err != nil {
		return err
	}

	user, err := u.db.GetUser(action.UserId)
	if err != nil {
		return err
	}
	if user == nil {
		return model.NewServiceError("invalid password token")
	}
	user.PasswordHash = Hash(*password)
	err = u.db.UpdateUser(user)
	if err != nil {
		return err
	}
	err = u.usersMail.SendSetPassword(user.Email)
	if err != nil {
		return err
	}
	err = u.actions.DeleteAction(action.Id)
	return err
}

func Hash(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}
