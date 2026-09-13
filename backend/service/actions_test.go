package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/syncloud/redirect/model"
	"testing"
	"time"
)

type ActionsDbStub struct {
	action *model.Action
}

func (db *ActionsDbStub) GetAction(_ int64, _ uint64) (*model.Action, error) {
	return db.action, nil
}
func (db *ActionsDbStub) GetActionByToken(_ string, _ uint64) (*model.Action, error) {
	return db.action, nil
}

func (db *ActionsDbStub) InsertAction(action *model.Action) error {
	db.action = action
	return nil
}

func (db *ActionsDbStub) UpdateAction(action *model.Action) error {
	if db.action != nil {
		db.action = action
	}
	return nil
}

func (db *ActionsDbStub) DeleteActions(_ int64) error {
	db.action = nil
	return nil
}

func (db *ActionsDbStub) DeleteAction(actionId uint64) error {
	db.action = nil
	return nil
}

type ClockStub struct {
	now time.Time
}

func (c *ClockStub) Now() time.Time {
	return c.now
}

func TestUpsert(t *testing.T) {

	db := &ActionsDbStub{nil}
	actions := NewActions(db, &ClockStub{now: time.Now()}, time.Hour)

	user := &model.User{Id: 1, Email: "test@example.com", PasswordHash: "pass", Active: true, UpdateToken: "token", Timestamp: time.Now()}
	action, err := actions.UpsertActivateAction(user.Id)

	assert.Nil(t, err)
	assert.NotNil(t, action)
	assert.NotNil(t, db.action)
}

func TestUpsert_New_CreatedNow(t *testing.T) {
	now := time.Now()
	db := &ActionsDbStub{nil}
	actions := NewActions(db, &ClockStub{now: now}, time.Hour)

	action, err := actions.UpsertPasswordAction(1)

	assert.NoError(t, err)
	assert.Equal(t, now, action.CreatedAt)
}

func TestUpsert_Existing_ResetsCreatedAt(t *testing.T) {
	now := time.Now()
	existing := &model.Action{
		Id:           1,
		ActionTypeId: ActionPassword,
		UserId:       1,
		Token:        "old",
		Timestamp:    now.Add(-10 * 24 * time.Hour),
		CreatedAt:    now.Add(-10 * 24 * time.Hour),
	}
	db := &ActionsDbStub{action: existing}
	actions := NewActions(db, &ClockStub{now: now}, time.Hour)

	action, err := actions.UpsertPasswordAction(1)

	assert.NoError(t, err)
	assert.NotEqual(t, "old", action.Token)
	assert.Equal(t, now, action.CreatedAt)
	assert.Equal(t, now, db.action.CreatedAt)
}

func TestGetPasswordAction_Fresh(t *testing.T) {
	now := time.Now()
	db := &ActionsDbStub{action: &model.Action{
		Id: 1, ActionTypeId: ActionPassword, UserId: 1, Token: "token", CreatedAt: now.Add(-30 * time.Minute),
	}}
	actions := NewActions(db, &ClockStub{now: now}, time.Hour)

	action, err := actions.GetPasswordAction("token")

	assert.NoError(t, err)
	assert.NotNil(t, action)
}

func TestGetPasswordAction_Expired(t *testing.T) {
	now := time.Now()
	db := &ActionsDbStub{action: &model.Action{
		Id: 1, ActionTypeId: ActionPassword, UserId: 1, Token: "token", CreatedAt: now.Add(-2 * time.Hour),
	}}
	actions := NewActions(db, &ClockStub{now: now}, time.Hour)

	action, err := actions.GetPasswordAction("token")

	assert.Error(t, err)
	assert.Nil(t, action)
}

func TestGetActivateAction_Old_StillValid(t *testing.T) {
	now := time.Now()
	db := &ActionsDbStub{action: &model.Action{
		Id: 1, ActionTypeId: ActionActivate, UserId: 1, Token: "token", CreatedAt: now.Add(-10000 * time.Hour),
	}}
	actions := NewActions(db, &ClockStub{now: now}, time.Hour)

	action, err := actions.GetActivateAction("token")

	assert.NoError(t, err)
	assert.NotNil(t, action)
}
