import smtplib
from email.mime.text import MIMEText


class Mail:
    def __init__(self, smtp_host, smtp_port, domain, email_from):
        self.email_from = email_from
        self.domain = domain
        self.smtp_host = smtp_host
        self.smtp_port = smtp_port

    def send(self, user_domain, email_to, activate_url):

        msg = MIMEText("""
        Hello

        You recently registered domain name at {0}, if this information is correct use the  link to activate it.

        Domain name: {1}.{0}

        Use the link to activate your domain: {2}
        """.format(self.domain, user_domain, activate_url))

        msg['Subject'] = 'Activate your domain: {0}.{1}'.format(user_domain, self.domain)
        msg['From'] = self.email_from
        msg['To'] = email_to

        s = smtplib.SMTP(self.smtp_host, self.smtp_port)
        s.sendmail(self.email_from, [email_to], msg.as_string())
        s.quit()