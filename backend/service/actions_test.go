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

func TestUpsert(t *testing.T) {

	db := &ActionsDbStub{nil}
	actions := NewActions(db)

	user := &model.User{Id: 1, Email: "test@example.com", PasswordHash: "pass", Active: true, UpdateToken: "token", PremiumStatusId: PremiumStatusInactive, Timestamp: time.Now()}
	action, err := actions.UpsertActivateAction(user.Id)

	assert.Nil(t, err)
	assert.NotNil(t, action)
	assert.NotNil(t, db.action)
}
