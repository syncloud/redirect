import smtplib
from email.mime.text import MIMEText


class Mail:
    def __init__(self, smtp_host, smtp_port, email_from, activate_url_template, password_url_template):
        self.smtp_host = smtp_host
        self.smtp_port = smtp_port
        self.email_from = email_from
        self.activate_url_template = activate_url_template
        self.password_url_template = password_url_template

    def send_activate(self, user_domain, domain, email_to, token):
        url = self.activate_url_template.format(token)

        if user_domain:
            full_domain = '{0}.{1}'.format(user_domain, domain)

            msg = MIMEText("""
            Hello,

            You recently registered domain name {0}, if this information is correct use the  link to activate it.

            Domain name: {0}

            Use the link to activate your account: {1}
            """.format(full_domain, url))
        else:
            msg = MIMEText("""
            Hello,

            You recently registered at {0}, if this information is correct use the link to activate it.

            Use the link to activate your account: {1}
            """.format(domain, url))


        msg['Subject'] = 'Activate account'
        msg['From'] = self.email_from
        msg['To'] = email_to

        s = smtplib.SMTP(self.smtp_host, self.smtp_port)
        s.sendmail(self.email_from, [email_to], msg.as_string())
        s.quit()

    def send_reset_password(self, email_to, token):
        url = self.password_url_template.format(token)

        msg = MIMEText("""
        Hello,

        The request to change your password was recently made.

        Use this link to reset your password: {0}
        """.format(url))

        msg['Subject'] = 'Reset password'
        msg['From'] = self.email_from
        msg['To'] = email_to

        s = smtplib.SMTP(self.smtp_host, self.smtp_port)
        s.sendmail(self.email_from, [email_to], msg.as_string())
        s.quit()