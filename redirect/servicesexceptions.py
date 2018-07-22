
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


def parameters_error(parameters_errors, message="There's an error in parameters"):
    return ParametersException(message, parameters_errors)


def parameter_error(parameter, error, message="There's an error in parameters"):
    return ParametersException(message, {parameter: [error]})
