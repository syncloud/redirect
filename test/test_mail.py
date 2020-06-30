import unittest
import smtp

from redirect.mail import Smtp, Mail


class TestMail(unittest.TestCase):
    smtp_host = 'mail'
    smtp_port = 1025

    def tearDown(self):
        smtp.clear()

    def test_send_log(self):
        mail = Mail(Smtp(self.smtp_host, self.smtp_port), 'support@redirect.com', None, None, 'support@redirect.com')
        logs = 'error logs'
        mail.send_logs('boris@email.com', logs, False)

        self.assertFalse(len(smtp.emails()) == 0)
        sent_mails = smtp.email_bodies()
        self.assertEquals(1, len(sent_mails))
        self.assertTrue(logs in sent_mails[0])
