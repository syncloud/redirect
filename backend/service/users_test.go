package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"testing"
	"time"
)

type UsersDbStub struct {
	user   *model.User
	action *model.Action
}

func (db *UsersDbStub) GetUserByUpdateToken(updateToken string) (*model.User, error) {
	panic("implement me")
}

func (db *UsersDbStub) GetUserByEmail(_ string) (*model.User, error) {
	return db.user, nil

}

func (db *UsersDbStub) GetUser(_ int64) (*model.User, error) {
	return db.user, nil
}

func (db *UsersDbStub) UpdateUser(user *model.User) error {
	db.user = user
	return nil
}

func (db *UsersDbStub) InsertUser(user *model.User) (int64, error) {
	db.user = user
	return user.Id, nil
}

func (db *UsersDbStub) DeleteUser(_ int64) error {
	return nil
}

type UsersActionsStub struct {
	action *model.Action
}

func (a *UsersActionsStub) GetActivateAction(_ string) (*model.Action, error) {
	return a.action, nil
}
func (a *UsersActionsStub) UpsertActivateAction(userId int64) (*model.Action, error) {
	a.action = &model.Action{Id: 1, ActionTypeId: ActionActivate, UserId: userId, Token: "token", Timestamp: time.Now()}
	return a.action, nil
}
func (a *UsersActionsStub) DeleteActions(_ int64) error {
	a.action = nil
	return nil
}

type UsersMailStub struct {
	sentEmail *string
	sentToken *string
}

func (a *UsersMailStub) SendActivate(to string, token string) error {
	a.sentEmail = &to
	a.sentToken = &token
	return nil
}

func (a *UsersMailStub) SendPremiumRequest(to string) error {
	a.sentEmail = &to
	return nil
}

var _ UsersMail = (*UsersMailStub)(nil)

func TestActivateSuccess(t *testing.T) {
	db := &UsersDbStub{}
	user := &model.User{Email: "test@example.com", PasswordHash: "password", Active: false, UpdateToken: "update token", PremiumStatusId: PremiumStatusInactive, Timestamp: time.Now()}
	actions := &UsersActionsStub{}
	action, _ := actions.UpsertActivateAction(user.Id)
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	_ = users.Save(user)

	err := users.Activate(action.Token)
	assert.Nil(t, err)
	assert.True(t, db.user.Active)
}

func TestActivateAlreadyActive(t *testing.T) {
	db := &UsersDbStub{}
	user := &model.User{Email: "test@example.com", PasswordHash: "password", Active: true, UpdateToken: "update token", PremiumStatusId: PremiumStatusInactive, Timestamp: time.Now()}
	actions := &UsersActionsStub{}
	action, _ := actions.UpsertActivateAction(user.Id)
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	_ = users.Save(user)

	err := users.Activate(action.Token)
	assert.NotNil(t, err)
}

func TestActivateWrongToken(t *testing.T) {
	db := &UsersDbStub{}
	user := &model.User{Email: "test@example.com", PasswordHash: "password", Active: true, UpdateToken: "update token", PremiumStatusId: PremiumStatusInactive, Timestamp: time.Now()}
	actions := &UsersActionsStub{}
	action, _ := actions.UpsertActivateAction(user.Id)
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	_ = users.Save(user)

	err := users.Activate(action.Token)
	assert.NotNil(t, err)
}

func TestActivateMissingUser(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	action, _ := actions.UpsertActivateAction(1)
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}

	err := users.Activate(action.Token)
	assert.NotNil(t, err)
}

func TestUserCreate(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, true, actions, mail}

	email := "test@example.com"
	password := "password"
	user, err := users.CreateNewUser(model.UserCreateRequest{Email: &email, Password: &password})
	assert.Nil(t, err)
	assert.Equal(t, email, user.Email)
	assert.NotEqual(t, password, user.PasswordHash)
	assert.Equal(t, email, user.Email)
	assert.False(t, user.Active)
	assert.Equal(t, user.Email, *mail.sentEmail)
	assert.Equal(t, actions.action.Token, *mail.sentToken)
}

func TestUserCreateSuccessNoActivation(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}

	email := "test@example.com"
	password := "password"
	user, err := users.CreateNewUser(model.UserCreateRequest{Email: &email, Password: &password})
	assert.Nil(t, err)
	assert.Equal(t, email, user.Email)
	assert.NotEqual(t, password, user.PasswordHash)
	assert.Equal(t, email, user.Email)

}

func TestUserCreateExistingEmail(t *testing.T) {
	db := &UsersDbStub{}
	user := &model.User{Email: "test@example.com", PasswordHash: "password", Active: true, UpdateToken: "update token", PremiumStatusId: PremiumStatusInactive, Timestamp: time.Now()}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	_ = users.Save(user)
	password := "password"
	user, err := users.CreateNewUser(model.UserCreateRequest{Email: &user.Email, Password: &password})
	assert.NotNil(t, err)
}

func TestUserMissingEmail(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}

	password := "password"
	_, err := users.CreateNewUser(model.UserCreateRequest{Email: nil, Password: &password})
	assert.NotNil(t, err)
	paramError := err.(*model.ParameterError)
	assert.NotNil(t, (*paramError.ParameterErrors)[0].Messages)

}

func TestPremiumRequest(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	user := &model.User{Email: "test@example.com", PasswordHash: "password", Active: true, UpdateToken: "update token", PremiumStatusId: PremiumStatusInactive, Timestamp: time.Now()}
	_ = users.Save(user)
	err := users.RequestPremiumAccount(user)
	assert.Nil(t, err)
	assert.Equal(t, PremiumStatusPending, user.PremiumStatusId)
	assert.Equal(t, user.Email, *mail.sentEmail)
}

func TestPremiumRequestAlreadyRequested(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	user := &model.User{Email: "test@example.com", PasswordHash: "password", Active: true, UpdateToken: "update token", PremiumStatusId: PremiumStatusPending, Timestamp: time.Now()}
	_ = users.Save(user)
	err := users.RequestPremiumAccount(user)
	assert.NotNil(t, err)
	assert.Equal(t, PremiumStatusPending, user.PremiumStatusId)
	assert.Nil(t, mail.sentEmail)
}
