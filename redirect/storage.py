import logging

from sqlalchemy import create_engine
from sqlalchemy import func
from sqlalchemy.orm import sessionmaker

import util
from models import User, Base, Action


class Storage:

    def __init__(self, session):
        self.session = session

    def get_user_by_email(self, email):
        user = self.session.query(User).filter(func.lower(User.email) == func.lower(email)).first()
        if user and not user.update_token:
            user.update_token = util.create_token()
        return user

    def add(self, *args):
        if len(args) > 0 and isinstance(args[0], list):
            args = args[0]
        for obj in args:
            self.session.add(obj)

    def clear(self):
        self.session.query(Action).delete()
        self.session.query(User).delete()


class SessionContext:

    def __init__(self, session):
        self.session = session

    def __enter__(self):
        return Storage(self.session)

    def __exit__(self, exc_type, exc_val, exc_tb):
        try:
            if exc_val is not None:
                logging.error('exception happened', exc_info=(exc_type, exc_val, exc_tb))
                raise exc_val
            else:
                try:
                    self.session.commit()
                except Exception, e:
                    logging.exception('unable to commit transaction')
                    self.session.rollback()
                    raise e
        finally:
            self.session.expunge_all()
            self.session.close()


class SessionContextFactory:

    def __init__(self, maker):
        self.maker = maker

    def __call__(self):
        return SessionContext(self.maker())


def get_session_maker(database_spec):
    engine = create_engine(database_spec, pool_pre_ping=True)
    maker = sessionmaker(expire_on_commit=False)
    maker.configure(bind=engine)
    Base.metadata.create_all(engine)
    return maker


def mysql_spec(host, user, password, database, port=3306):
    database_spec = "mysql+mysqldb://{0}:{1}@{2}:{3}/{4}".format(user, password, host, port, database)
    return database_spec


def mysql_spec_config(config):
    host = config.get('mysql', 'host')
    user = config.get('mysql', 'user')
    password = config.get('mysql', 'passwd')
    database = config.get('mysql', 'db')
    return mysql_spec(host, user, password, database)
