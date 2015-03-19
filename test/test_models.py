import unittest
from redirect.models import Service, new_service


class TestService(unittest.TestCase):

    def test_service__str__(self):
        service = new_service("service", "type", 80)
        self.assertEqual("name: service, type: type, port: 80", service.__str__())