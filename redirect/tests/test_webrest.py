import unittest
import requests
import uuid
from urlparse import urljoin

class TestWebRest(unittest.TestCase):

    base_url = r'http://127.0.0.1:5000'

    def post(self, url, params):
        return requests.post(urljoin(self.base_url, url), data=params)

    def test_user_create_success(self):
        user_domain = uuid.uuid4().hex
        email = user_domain+'@mail.com'
        params = {'user_domain': user_domain, 'email': email, 'password': 'pass123456'}
        response = self.post('user/create', params)
        self.assertTrue(response.ok)
        self.assertEqual(200, response.status_code)
#        token = response.headers['token']
#        self.assertIsNotNone(token)
#        self.assertNotEqual('', token)

if __name__ == '__main__':
    unittest.run()
