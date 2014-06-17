from sqlalchemy import Table, Column, Integer, String, Boolean, ForeignKey
from sqlalchemy.orm import relationship, backref
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()

class User(Base):
    __tablename__ = "user"
    id = Column(Integer, primary_key=True)
    email = Column(String())
    password_hash = Column(String())
    active = Column(Boolean())
    activate_token = Column(String())

    domains = relationship("Domain", lazy='subquery')

    def __init__(self, email, password_hash, active, activate_token):
        self.email = email
        self.password_hash = password_hash
        self.active = active
        self.activate_token = activate_token

    def update_active(self, active):
        self.active = active


class Domain(Base):
    __tablename__ = "domain"
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
    id = Column(Integer, primary_key=True)
    name = Column(String())
    type = Column(String())
    port = Column(Integer())

    domain_id = Column(Integer, ForeignKey('domain.id'))
    domain = relationship("Domain", lazy='subquery')

    def __init__(self, name, type, port):
        self.name = name
        self.type = type
        self.port = port