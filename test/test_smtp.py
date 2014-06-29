import unittest
import smtplib
from fakesmtp import FakeSmtp
from email.mime.text import MIMEText
from redirect.mail import Mail

class TestSmtp(unittest.TestCase):
    smtp_outbox_path = 'outbox'
    smtp_host = 'localhost'
    smtp_port = 2500

    def setUp(self):
        self.smtp = FakeSmtp(self.smtp_outbox_path)
        self.smtp.clear()

    def test_send_goes_to_file(self):
        self.assertTrue(self.smtp.empty())

        email_from = 'from@mail.com'
        email_to = 'to@mail.com'

        msg = MIMEText('Text message should be here')
        msg['Subject'] = 'Some subject'
        msg['From'] = email_from
        msg['To'] = email_to

        s = smtplib.SMTP(self.smtp_host, self.smtp_port)
        s.sendmail(email_from, [email_to], msg.as_string())
        s.quit()

        self.assertFalse(self.smtp.empty())
        sent_mails = self.smtp.emails()
        self.assertEquals(1, len(sent_mails))

    def test_mail_send(self):
        activate_url = 'http://redirect.com/activate?token=t123456'
        mail = Mail(self.smtp_host, self.smtp_port, 'support@redirect.com')
        mail.send_activate('boris', 'redirect.com', 'boris@email.com', activate_url)

        self.assertFalse(self.smtp.empty())
        sent_mails = self.smtp.emails()
        self.assertEquals(1, len(sent_mails))
        self.assertTrue(activate_url in sent_mails[0])
