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
	user := db.user
	if user != nil {
		//copy
		user = &model.User{
			Id:                  db.user.Id,
			Email:               db.user.Email,
			PasswordHash:        db.user.PasswordHash,
			Active:              db.user.Active,
			UpdateToken:         db.user.UpdateToken,
			NotificationEnabled: db.user.NotificationEnabled,
			PremiumStatusId:     db.user.PremiumStatusId,
			Timestamp:           db.user.Timestamp,
		}
	}
	return user, nil
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

func (a *UsersActionsStub) SendSetPassword(to string) error {
	panic("implement me")
}

func (a *UsersActionsStub) GetPasswordAction(token string) (*model.Action, error) {
	return a.action, nil
}

func (a *UsersActionsStub) DeleteAction(actionId uint64) error {
	a.action = nil
	return nil
}

func (a *UsersActionsStub) UpsertPasswordAction(userId int64) (*model.Action, error) {
	a.action = &model.Action{Id: 1, ActionTypeId: ActionPassword, UserId: userId, Token: "token", Timestamp: time.Now()}
	return a.action, nil
}

type UsersMailStub struct {
	sentEmail *string
	sentToken *string
}

func (a *UsersMailStub) SendSetPassword(to string) error {
	return nil
}

func (a *UsersMailStub) SendActivate(to string, token string) error {
	a.sentEmail = &to
	a.sentToken = &token
	return nil
}

func (a *UsersMailStub) SendPlanSubscribed(to string) error {
	a.sentEmail = &to
	return nil
}

func (a *UsersMailStub) SendResetPassword(to string, token string) error {
	return nil
}

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

func TestUsers_PlanSubscribe_Good(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	user := &model.User{Email: "test@example.com", PasswordHash: "password", Active: true, UpdateToken: "update token", Timestamp: time.Now()}
	_ = users.Save(user)
	err := users.PlanSubscribe(user, "123")
	assert.Nil(t, err)
	assert.Equal(t, "123", *user.SubscriptionId)
	assert.Equal(t, user.Email, *mail.sentEmail)
}

func TestUsers_PlanSubscribe_AlreadySubscribed(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	subscriptionId := "123"
	user := &model.User{Email: "test@example.com", PasswordHash: "password", Active: true, UpdateToken: "update token", SubscriptionId: &subscriptionId, Timestamp: time.Now()}
	_ = users.Save(user)
	err := users.PlanSubscribe(user, "123")
	assert.NotNil(t, err)
	assert.Equal(t, "123", *user.SubscriptionId)
	assert.Nil(t, mail.sentEmail)
}

func TestUserAuthenticateSuccess(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	email := "test@example.com"
	password := "password"
	user := &model.User{Email: email, PasswordHash: Hash(password), Active: true, UpdateToken: "update token", PremiumStatusId: PremiumStatusPending, Timestamp: time.Now()}
	_ = users.Save(user)
	authenticatedUser, err := users.Authenticate(&email, &password)

	assert.Nil(t, err)
	assert.NotNil(t, authenticatedUser)
}

func TestUserAuthenticateWrongPassword(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	email := "test@example.com"
	user := &model.User{Email: email, PasswordHash: Hash("otherpassword"), Active: true, UpdateToken: "update token", PremiumStatusId: PremiumStatusPending, Timestamp: time.Now()}
	_ = users.Save(user)
	password := "password"
	_, err := users.Authenticate(&email, &password)

	assert.NotNil(t, err)
}

func TestUserAuthenticateNotExisting(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	email := "test@example.com"
	password := "password"
	_, err := users.Authenticate(&email, &password)

	assert.NotNil(t, err)

}

func TestUserAuthenticateNonActive(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	email := "test@example.com"
	user := &model.User{Email: email, PasswordHash: Hash("otherpassword"), Active: false, UpdateToken: "update token", PremiumStatusId: PremiumStatusPending, Timestamp: time.Now()}
	_ = users.Save(user)
	password := "password"
	_, err := users.Authenticate(&email, &password)

	assert.NotNil(t, err)
}

func TestUserAuthenticateMissingPassword(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	email := "test@example.com"
	user := &model.User{Email: email, PasswordHash: Hash("otherpassword"), Active: false, UpdateToken: "update token", PremiumStatusId: PremiumStatusPending, Timestamp: time.Now()}
	_ = users.Save(user)
	_, err := users.Authenticate(&email, nil)

	assert.NotNil(t, err)
}

func TestPasswordReset(t *testing.T) {
	db := &UsersDbStub{}
	actions := &UsersActionsStub{}
	mail := &UsersMailStub{}
	users := &Users{db, false, actions, mail}
	email := "test@example.com"
	password1 := "password1"
	user := &model.User{Email: email, PasswordHash: Hash(password1), Active: true, UpdateToken: "update token", PremiumStatusId: PremiumStatusPending, Timestamp: time.Now()}
	_ = users.Save(user)
	_, err := users.Authenticate(&email, &password1)
	assert.Nil(t, err)
	token, err := users.RequestPasswordReset(email)
	assert.Nil(t, err)
	password2 := "password2"
	request := &model.UserPasswordSetRequest{Token: token, Password: &password2}
	err = users.UserSetPassword(request)
	assert.Nil(t, err)
	_, err = users.Authenticate(&email, &password2)
	assert.Nil(t, err)
}

func TestHashNotEmpty(t *testing.T) {
	hash := Hash("non empty string")
	assert.NotEmpty(t, hash)
}

func TestEqualInput(t *testing.T) {
	h1 := Hash("some string")
	h2 := Hash("some string")
	assert.Equal(t, h1, h2)
}

func testNotEqualInput(t *testing.T) {
	h1 := Hash("some string")
	h2 := Hash("some other string")
	assert.NotEqual(t, h1, h2)
}
