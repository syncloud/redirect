from sqlalchemy import Table, Column, Integer, String, Boolean, ForeignKey, DateTime
from sqlalchemy.orm import relationship, backref
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()


def limit(dictionary, keys):
    filtered = [key for key in dictionary.keys() if key not in keys]
    for key in filtered:
        dictionary.pop(key, None)

def to_dict(value):
    if isinstance(value, list):
        return [to_dict(item) for item in value]
    if isinstance(value, dict):
        cloned = value.copy()
        for key, value in cloned.items():
            cloned[key] = to_dict(value)
        return cloned
    if value is None or isinstance(value, (bool, basestring, str, unicode, int, long, float)):
        return value
    result = value.__dict__.copy()
    if hasattr(value, '__public__'):
        limit(result, value.__public__)
    for member, mvalue in result.items():
        result[member] = to_dict(mvalue)
    return result


def fromdict(self, values):
    for c in self.__table__.columns:
        if c.name in values:
            setattr(self, c.name, values[c.name])

Base.fromdict = fromdict


class User(Base):
    __tablename__ = "user"
    id = Column(Integer, primary_key=True)
    email = Column(String())
    password_hash = Column(String())
    active = Column(Boolean())
    update_token = Column(String())

    domains = relationship("Domain", lazy='subquery')
    actions = relationship("Action", lazy='subquery', cascade="all, delete, delete-orphan")

    def __init__(self, email, password_hash, active):
        self.email = email
        self.password_hash = password_hash
        self.active = active

    def update_active(self, active):
        self.active = active

    def set_activate_token(self, token):
        self.actions.append(Action(token, ActionType.ACTIVATE))

    def set_password_token(self, token):
        self.actions.append(Action(token, ActionType.ACTIVATE))

    def activate_token(self):
        return self.get_action_token(ActionType.ACTIVATE)

    def password_token(self):
        return self.get_action_token(ActionType.PASSWORD)

    def get_action_token(self, type):
        action = next(action for action in self.actions if action.action_type_id == type)
        if action:
            return action.token


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


class Domain(Base):
    __tablename__ = "domain"
    __public__ = ['user_domain', 'ip', 'services']

    id = Column(Integer, primary_key=True)
    user_domain = Column(String())
    ip = Column(String())
    update_token = Column(String())

    user_id = Column(Integer, ForeignKey('user.id'))
    user = relationship("User", lazy='subquery')

    services = relationship("Service", lazy='subquery')

    def __init__(self, user_domain, ip=None, update_token=None):
        self.user_domain = user_domain
        self.ip = ip
        self.update_token = update_token


class Service(Base):
    __tablename__ = "service"
    __public__ = ['name', 'type', 'port', 'url']

    id = Column(Integer, primary_key=True)
    name = Column(String())
    type = Column(String())
    url = Column(String())
    port = Column(Integer())

    domain_id = Column(Integer, ForeignKey('domain.id'))
    domain = relationship("Domain", lazy='subquery')


def new_service(name, type, port):
    s = Service()
    s.name = name
    s.type = type
    s.port = port
    return s


def new_service_fromdict(dictionary):
    s = Service()
    s.fromdict(dictionary)
    return s
