import os
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


def read_letter(filepath):
    f = open(filepath, 'r')
    subject_line = f.readline()
    subject = subject_line.replace('Subject:', '')
    subject = subject.strip()
    text = f.read()
    f.close()
    return subject, text


class Mail:
    def __init__(self, smtp, email_from, activate_url_template, password_url_template):
        self.smtp = smtp
        self.email_from = email_from
        self.activate_url_template = activate_url_template
        self.password_url_template = password_url_template
        self.path = 'emails'

    def email_path(self, filename):
        return os.path.join(self.path, filename)

    def send_activate(self, main_domain, email_to, token):
        url = self.activate_url_template.format(token)

        subject, letter = read_letter(self.email_path('activate.txt'))

        msg = MIMEText(letter.format(main_domain=main_domain, url=url))

        msg['Subject'] = subject
        msg['From'] = self.email_from
        msg['To'] = email_to

        self.smtp.send(self.email_from, email_to, msg.as_string())

    def send_reset_password(self, email_to, token):
        url = self.password_url_template.format(token)

        subject, letter = read_letter(self.email_path('reset_password.txt'))

        msg = MIMEText(letter.format(url=url))

        msg['Subject'] = subject
        msg['From'] = self.email_from
        msg['To'] = email_to

        self.smtp.send(self.email_from, email_to, msg.as_string())

    def send_set_password(self, email_to):
        subject, letter = read_letter(self.email_path('set_password.txt'))

        msg = MIMEText(letter)

        msg['Subject'] = subject
        msg['From'] = self.email_from
        msg['To'] = email_to

        self.smtp.send(self.email_from, email_to, msg.as_string())