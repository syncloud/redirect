import logging
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from models import User, Domain, Service, Base


class Storage:

    def __init__(self, session):
        self.session = session

    def delete_user(self, email):
        user = self.get_user_by_email(email)
        if user is not None:
            self.session.delete(user)
            return True
        return False

    def get_user_by_email(self, email):
        user = self.session.query(User).filter(User.email == email).first()
        return user

    def get_user_by_activate_token(self, activate_token):
        user = self.session.query(User).filter(User.activate_token == activate_token).first()
        return user

    def get_domain_by_update_token(self, update_token):
        domain = self.session.query(Domain).filter(Domain.update_token == update_token).first()
        return domain

    def get_domain_by_name(self, user_domain):
        domain = self.session.query(Domain).filter(Domain.user_domain == user_domain).first()
        return domain

    def add(self, *args):
        for obj in args:
            self.session.add(obj)

    def clear(self):
        self.session.query(Service).delete()
        self.session.query(Domain).delete()
        self.session.query(User).delete()


class SessionContext:

    def __init__(self, session):
        self.session = session

    def __enter__(self):
        return Storage(self.session)

    def __exit__(self, exc_type, exc_val, exc_tb):
        try:
            if exc_val is None:
                self.session.commit()
            else:
                logging.error('exception happened', exc_info=(exc_type, exc_val, exc_tb))
                raise exc_val
        except Exception, e:
            logging.exception('unable to commit transaction', e)
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
    engine = create_engine(database_spec)
    maker = sessionmaker(expire_on_commit = False)
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
