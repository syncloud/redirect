from sqlalchemy import Table, Column, Integer, String, Boolean, ForeignKey, DateTime
from sqlalchemy.orm import relationship, backref
from sqlalchemy.ext.declarative import declarative_base

import util

Base = declarative_base()

def from_dict(self, values):
    for c in self.__table__.columns:
        if c.name in values:
            setattr(self, c.name, values[c.name])

Base.from_dict = from_dict


class User(Base):
    __tablename__ = "user"
    __public__ = ['email', 'active', 'unsubscribed', 'domains', 'update_token']

    id = Column(Integer, primary_key=True)
    email = Column(String())
    password_hash = Column(String())
    active = Column(Boolean())
    update_token = Column(String())
    unsubscribed = Column(Boolean())

    domains = relationship("Domain", lazy='subquery')
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


class Domain(Base):
    __tablename__ = "domain"
    __public__ = ['user_domain', 'ip', 'local_ip', 'map_local_address', 'device_mac_address', 'device_name', 'device_title', 'services', 'last_update']

    id = Column(Integer, primary_key=True)
    user_domain = Column(String())
    ip = Column(String())
    local_ip = Column(String())
    map_local_address = Column(Boolean())
    update_token = Column(String())
    last_update = Column(DateTime)
    device_mac_address = Column(String())
    device_name = Column(String())
    device_title = Column(String())
    user_id = Column(Integer, ForeignKey('user.id'))
    user = relationship("User", lazy='subquery')

    services = relationship("Service", lazy='subquery')

    def __init__(self, user_domain, device_mac_address, device_name, device_title, update_token):
        self.user_domain = user_domain
        self.update_token = update_token
        self.device_mac_address = device_mac_address
        self.device_name = device_name
        self.device_title = device_title

    def dns_name(self, main_domain):
        return '{0}.{1}.'.format(self.user_domain, main_domain)

    def dns_ip(self):
        if self.map_local_address:
            return self.local_ip
        return self.ip


class Service(Base):
    __tablename__ = "service"
    __public__ = ['name', 'protocol', 'type', 'port', 'local_port', 'url']

    id = Column(Integer, primary_key=True)
    name = Column(String())
    protocol = Column(String())
    type = Column(String())
    url = Column(String())
    port = Column(Integer())
    local_port = Column(Integer())

    domain_id = Column(Integer, ForeignKey('domain.id'))
    domain = relationship("Domain", lazy='subquery')

    def dns_name(self, main_domain):
        return '{0}.{1}.{2}.'.format(self.type, self.domain.user_domain, main_domain)

    def dns_value(self, main_domain):
        return '0 0 {0} {1}.{2}.'.format(self.dns_port(), self.domain.user_domain, main_domain)

    def dns_port(self):
        if self.domain.map_local_address:
            return self.local_port
        return self.port

    def __str__(self):
        return "{ " + ", ".join(["{0}: {1}".format(f, getattr(self, f)) for f in self.fields_str()]) + " }"

    def fields_str(self):
        return [field for field in self.__public__ if getattr(self, field)]

    def __repr__(self):
        return self.__str__()

def new_service(name, type, port):
    s = Service()
    s.name = name
    s.type = type
    s.port = port
    return s


def new_service_from_dict(dictionary):
    s = Service()
    s.from_dict(dictionary)
    return s
