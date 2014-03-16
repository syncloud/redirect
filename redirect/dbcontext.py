import logging


class DbContext:

    def __init__(self, connection):
        self.connection = connection
        self.cursor = connection.cursor()

    def __enter__(self):
        return self.cursor

    def __exit__(self, exc_type, exc_val, exc_tb):
        try:
            self.cursor.close()
        except Exception, e:
            logging.error("unable to close cursor: {0}".format(e.message))

        if isinstance(exc_val, Exception):
            self.connection.rollback()
        else:
            self.connection.commit()

        try:
            self.connection.close()
        except Exception, e:
            logging.error("unable to close connection: {0}".format(e.message))