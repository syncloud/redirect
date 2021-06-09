import unittest

from redirect.util import create_token


class TestCreateToken(unittest.TestCase):

    def test_length(self):
        token = create_token()
        self.assertIsNotNone(token)
        self.assertTrue(len(token) > 10)

    def test_different_every_time(self):
        token1 = create_token()
        token2 = create_token()
        self.assertNotEquals(token1, token2)


if __name__ == '__main__':
    unittest.run()
