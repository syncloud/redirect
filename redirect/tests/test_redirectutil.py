import unittest
from .. redirectutil import hash

class TestHash(unittest.TestCase):

    def test_empty(self):
        h = hash('non empty string')
        self.assertIsNotNone(h)
        self.assertIsNot('', h)

    def test_equal_input(self):
        h1 = hash('some string')
        h2 = hash('some string')
        self.assertEqual(h1, h2)

    def test_not_equal_input(self):
        h1 = hash('some string')
        h2 = hash('some other string')
        self.assertNotEqual(h1, h2)

if __name__ == '__main__':
    unittest.run()