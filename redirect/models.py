from sqlalchemy import Column, Integer, String, Boolean, ForeignKey
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()

def from_dict(self, values):
    for c in self.__table__.columns:
        if c.name in values:
            setattr(self, c.name, values[c.name])

Base.from_dict = from_dict


class User(Base):
    __tablename__ = "user"
    __public__ = ['email', 'active', 'update_token']

    id = Column(Integer, primary_key=True)
    email = Column(String())
    active = Column(Boolean())
    update_token = Column(String())
