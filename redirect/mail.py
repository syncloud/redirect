import smtplib
from email.mime.text import MIMEText


class Mail:
    def __init__(self, smtp_host, smtp_port, email_from):
        self.email_from = email_from
        self.smtp_host = smtp_host
        self.smtp_port = smtp_port

    def send_activate(self, full_domain, email_to, activate_url):

        msg = MIMEText("""
        Hello

        You recently registered domain name {0}, if this information is correct use the  link to activate it.

        Domain name: {0}

        Use the link to activate your domain: {1}
        """.format(full_domain, activate_url))

        msg['Subject'] = 'Activate your domain: {0}'.format(full_domain)
        msg['From'] = self.email_from
        msg['To'] = email_to

        s = smtplib.SMTP(self.smtp_host, self.smtp_port)
        s.sendmail(self.email_from, [email_to], msg.as_string())
        s.quit()