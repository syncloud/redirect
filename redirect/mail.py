import smtplib
from email.mime.text import MIMEText


class Mail:
    def __init__(self, domain, email_from, api_url):
        self.api_url = api_url
        self.email_from = email_from
        self.domain = domain

    def send(self, user_domain, email_to, token):

        msg = MIMEText("""
        Hello

        You recently registered domain name at {0}, if this information is correct use the  link to activate it.

        Domain name: {1}.{0}

        Use the link to activate your domain: http://{2}/activate?token={3}
        """.format(self.domain, user_domain, self.api_url, token))

        msg['Subject'] = 'Activate your domain: {0}.{1}'.format(user_domain, self.domain)
        msg['From'] = self.email_from
        msg['To'] = email_to

        s = smtplib.SMTP('localhost')
        s.sendmail(self.email_from, [email_to], msg.as_string())
        s.quit()