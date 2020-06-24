import unittest
from fakesmtp import FakeSmtp

from redirect.mail import Smtp, Mail


class TestMail(unittest.TestCase):
    smtp_host = 'localhost'
    smtp_port = 2500

    def setUp(self):
        self.smtp = FakeSmtp(self.smtp_host, self.smtp_port)

    def tearDown(self):
        self.smtp.stop()

    def test_send_log(self):
        mail = Mail(Smtp(self.smtp_host, self.smtp_port), 'support@redirect.com', None, None, 'support@redirect.com')
        logs = 'error logs'
        mail.send_logs('boris@email.com', logs, False)

        self.assertFalse(self.smtp.empty())
        sent_mails = self.smtp.emails()
        self.assertEquals(1, len(sent_mails))
        self.assertTrue(logs in sent_mails[0])
