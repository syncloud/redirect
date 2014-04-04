import unittest
import shutil
import os
import smtplib
from email.mime.text import MIMEText

class TestSmtp(unittest.TestCase):
    smtp_outbox_path = 'outbox'
    smtp_host = 'localhost'
    smtp_port = 2500

    def setUp(self):
        shutil.rmtree(self.smtp_outbox_path)

    def test_send_goes_to_file(self):
        email_from = 'from@mail.com'
        email_to = 'to@mail.com'

        msg = MIMEText('Text message should be here')
        msg['Subject'] = 'Some subject'
        msg['From'] = email_from
        msg['To'] = email_to

        s = smtplib.SMTP(self.smtp_host, self.smtp_port)
        s.sendmail(email_from, [email_to], msg.as_string())
        s.quit()

        sent_mails = os.listdir(self.smtp_outbox_path)
        self.assertEqual(1, len(sent_mails))

