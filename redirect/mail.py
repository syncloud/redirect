import smtplib
from email.mime.text import MIMEText


class Smtp:
    def __init__(self, smtp_host, smtp_port, use_tls=False, login=None, password=None):
        self.smtp_host = smtp_host
        self.smtp_port = smtp_port
        self.use_tls = use_tls
        self.login = login
        self.password = password

    def send(self, email_from, email_to, msg_string):
        s = smtplib.SMTP()
        s.connect(self.smtp_host, self.smtp_port)
        s.ehlo()
        if self.use_tls:
            s.starttls()
        if self.login:
            s.login(self.login, self.password)
        s.sendmail(email_from, [email_to], msg_string)
        s.quit()


class Mail:
    def __init__(self, smtp, email_from, activate_url_template, password_url_template):
        self.smtp = smtp
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

        self.smtp.send(self.email_from, email_to, msg.as_string())

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

        self.smtp.send(self.email_from, email_to, msg.as_string())

    def send_set_password(self, email_to):
        msg = MIMEText("""
        Hello,

        Your password has been reset.
        """)

        msg['Subject'] = 'Reset password'
        msg['From'] = self.email_from
        msg['To'] = email_to

        self.smtp.send(self.email_from, email_to, msg.as_string())