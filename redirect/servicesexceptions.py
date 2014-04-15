
class ServiceException(Exception):
    status_code = 400

    def __init__(self, message, status_code=None):
        Exception.__init__(self)
        self.message = message
        if status_code is not None:
            self.status_code = status_code


def bad_request(message):
    return ServiceException(message)


def conflict(message):
    return ServiceException(message, status_code=409)


def forbidden(message):
    return ServiceException(message, status_code=403)


def not_found(message):
    return ServiceException(message, status_code=404)
