import smtplib
from email.mime.text import MIMEText


class Mail:
    def __init__(self, domain, email_from):
        self.email_from = email_from
        self.domain = domain

    def send(self, username, email_to, token):

        msg = MIMEText("""
        Hello

        You recently registered domain name at {0}, if this information is correct use the  link to activate it.

        Domain name: {1}.{0}

        Use the link to activate your domain: http://{0}/activate?token={2}
        """.format(self.domain, username, token))

        msg['Subject'] = 'Activate your domain: {0}.{1}'.format(username, self.domain)
        msg['From'] = self.email_from
        msg['To'] = email_to

        s = smtplib.SMTP('localhost')
        s.sendmail(self.email_from, [email_to], msg.as_string())
        s.quit()