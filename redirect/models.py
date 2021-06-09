from sqlalchemy import Table, Column, Integer, String, Boolean, ForeignKey, DateTime
from sqlalchemy.orm import relationship, backref
from sqlalchemy.ext.declarative import declarative_base
from IPy import IP
import util

Base = declarative_base()

def from_dict(self, values):
    for c in self.__table__.columns:
        if c.name in values:
            setattr(self, c.name, values[c.name])

Base.from_dict = from_dict


class User(Base):
    __tablename__ = "user"
    __public__ = ['email', 'active', 'unsubscribed', 'update_token']

    id = Column(Integer, primary_key=True)
    email = Column(String())
    password_hash = Column(String())
    active = Column(Boolean())
    update_token = Column(String())
    unsubscribed = Column(Boolean())

    actions = relationship("Action", lazy='subquery', cascade="all, delete, delete-orphan")

    def __init__(self, email, password_hash, active):
        self.email = email
        self.password_hash = password_hash
        self.active = active
        self.unsubscribed = False
        self.update_token = util.create_token()

    def enable_action(self, type):
        token = util.create_token()
        action = self.action(type)
        if action:
            action.token = token
        else:
            action = Action(token, type)
            self.actions.append(action)
        return action

    def token(self, type):
        action = self.action(type)
        if action:
            return action.token

    def action(self, type):
        return next((action for action in self.actions if action.action_type_id == type), None)


class Action(Base):
    __tablename__ = "action"
    id = Column(Integer, primary_key=True)
    action_type_id = Column(Integer, ForeignKey('action_type.id'))
    user_id = Column(Integer, ForeignKey('user.id'))
    token = Column(String())

    user = relationship("User", lazy='subquery')
    action_type = relationship("ActionType", lazy='subquery')

    def __init__(self, token, action_type_id):
        self.token = token
        self.action_type_id = action_type_id


class ActionType(Base):
    ACTIVATE = 1
    PASSWORD = 2
    __tablename__ = "action_type"
    id = Column(Integer, primary_key=True)
    name = Column(String())

    def __init__(self, name):
        self.name = name

    def update_active(self, active):
        self.active = active
