import logging
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

import models

def mysql_spec(mysql_host, mysql_user, mysql_passwd, mysql_db):
    database_spec = "mysql+mysqldb://{0}:{1}@{2}:{3}/{4}".format(mysql_user, mysql_passwd, mysql_host, 3306, mysql_db)
    return database_spec

def get_session_maker(database_spec):
    engine = create_engine(database_spec)
    maker = sessionmaker(expire_on_commit = False)
    maker.configure(bind=engine)
    models.Base.metadata.create_all(engine)
    return maker

class SessionContext:

    def __init__(self, session):
        self.session = session

    def __enter__(self):
        return self.session

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