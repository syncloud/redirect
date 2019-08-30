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
    __public__ = ['user_domain', 'ip', 'ipv6', 'local_ip', 'map_local_address', 'device_mac_address', 'device_name', 'device_title', 'platform_version', 'web_protocol', 'web_port', 'web_local_port', 'last_update']

    id = Column(Integer, primary_key=True)
    user_domain = Column(String())
    ip = Column(String())
    ipv6 = Column(String())
    local_ip = Column(String())
    map_local_address = Column(Boolean())
    update_token = Column(String())
    last_update = Column(DateTime)
    device_mac_address = Column(String())
    device_name = Column(String())
    device_title = Column(String())
    platform_version = Column(String())
    web_protocol = Column(String())
    web_port = Column(Integer())
    web_local_port = Column(Integer())
    user_id = Column(Integer, ForeignKey('user.id'))
    user = relationship("User", lazy='subquery')

    def __init__(self, user_domain, device_mac_address, device_name, device_title, update_token):
        self.user_domain = user_domain
        self.update_token = update_token
        self.device_mac_address = device_mac_address
        self.device_name = device_name
        self.device_title = device_title

    def dns_name(self, main_domain):
        return '{0}.{1}.'.format(self.user_domain, main_domain)

    def dns_wildcard_name(self, main_domain):
        return '\\052.{0}.{1}.'.format(self.user_domain, main_domain)

    def access_ip(self):
        if self.map_local_address:
            return self.local_ip
        return self.ip


    def dns_ipv6(self):
        if self.ipv6 and IP(self.ipv6).version() == 6:
            return self.ipv6
        access_ip = self.access_ip()
        if access_ip and IP(access_ip).version() == 6:
            return access_ip
        return None
    
    def dns_ipv4(self):
        access_ip = self.access_ip()
        if access_ip and IP(access_ip).version() == 4:
            return access_ip
        return None

    def has_dns_ip(self):
        return self.dns_ipv6() is not None or self.dns_ipv4() is not None
