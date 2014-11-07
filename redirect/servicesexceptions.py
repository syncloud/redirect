
class ServiceException(Exception):
    def __init__(self, message, status_code=400):
        Exception.__init__(self)
        self.message = message
        if status_code is not None:
            self.status_code = status_code


class ParametersException(ServiceException):
    def __init__(self, message, parameters_errors, status_code=400):
        ServiceException.__init__(self, message, status_code)
        self.parameters_errors = parameters_errors


def bad_request(message):
    return ServiceException(message)


def conflict(message):
    return ServiceException(message, status_code=409)


def forbidden(message):
    return ServiceException(message, status_code=403)


def not_found(message):
    return ServiceException(message, status_code=404)


def parameters_error(parameters_errors, message="There's a error in parameters"):
    return ParametersException(message, parameters_errors)
