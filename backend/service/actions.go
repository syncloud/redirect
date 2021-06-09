package service

import (
	"fmt"
	"github.com/syncloud/redirect/model"
	"github.com/syncloud/redirect/utils"
	"time"
)

const (
	ActionActivate = 1
	ActionPassword = 2
)

type ActionsDb interface {
	GetAction(userId int64, actionTypeId uint64) (*model.Action, error)
	GetActionByToken(token string, actionTypeId uint64) (*model.Action, error)
	InsertAction(action *model.Action) error
	UpdateAction(action *model.Action) error
	DeleteActions(userId int64) error
	DeleteAction(actionId uint64) error
}

type Actions struct {
	db ActionsDb
}

func NewActions(db ActionsDb) *Actions {
	return &Actions{db: db}
}

func (a *Actions) GetActivateAction(token string) (*model.Action, error) {
	action, err := a.db.GetActionByToken(token, ActionActivate)
	if err != nil {
		return nil, err
	}
	if action == nil {
		return nil, &model.ServiceError{InternalError: fmt.Errorf("invalid activation token")}
	}
	return action, err
}

func (a *Actions) GetPasswordAction(token string) (*model.Action, error) {
	action, err := a.db.GetActionByToken(token, ActionPassword)
	if err != nil {
		return nil, err
	}
	if action == nil {
		return nil, &model.ServiceError{InternalError: fmt.Errorf("invalid password token")}
	}
	return action, err
}

func (a *Actions) UpsertActivateAction(userId int64) (*model.Action, error) {
	return a.upsertAction(userId, ActionActivate)
}

func (a *Actions) DeleteActions(userId int64) error {
	return a.db.DeleteActions(userId)
}

func (a *Actions) DeleteAction(actionId uint64) error {
	return a.db.DeleteAction(actionId)
}

func (a *Actions) UpsertPasswordAction(userId int64) (*model.Action, error) {
	return a.upsertAction(userId, ActionPassword)
}

func (a *Actions) upsertAction(userId int64, actionTypeId uint64) (*model.Action, error) {
	token := utils.Uuid()
	now := time.Now()
	action, err := a.db.GetAction(userId, actionTypeId)
	if err != nil {
		return nil, err
	}
	if action != nil {
		action.Token = token
		action.Timestamp = now
		err = a.db.UpdateAction(action)
		if err != nil {
			return nil, err
		}
	} else {
		action = &model.Action{
			ActionTypeId: actionTypeId, UserId: userId, Token: token, Timestamp: now,
		}
		err = a.db.InsertAction(action)
		if err != nil {
			return nil, err
		}
	}
	return action, nil
}
