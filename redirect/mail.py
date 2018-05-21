import os
import smtplib
from email.mime.text import MIMEText
from os.path import dirname, join, splitext
import tempfile


class Smtp:
    def __init__(self, smtp_host, smtp_port, use_tls=False, login=None, password=None):
        self.smtp_host = smtp_host
        self.smtp_port = smtp_port
        self.use_tls = use_tls
        self.login = login
        self.password = password

    def send(self, email_from, email_to_list, msg_string):
        s = smtplib.SMTP()
        s.connect(self.smtp_host, self.smtp_port)
        s.ehlo()
        if self.use_tls:
            s.starttls()
        if self.login:
            s.login(self.login, self.password)
        s.sendmail(email_from, email_to_list, msg_string)
        s.quit()


def get_smtp(config):
    smtp_host = config.get('smtp', 'host')
    smtp_port = config.getint('smtp', 'port')
    smtp_use_tls = False
    smtp_login = None
    smtp_password = None
    if config.has_option('smtp', 'use_tls'):
        smtp_use_tls = config.getboolean('smtp', 'use_tls')
    if config.has_option('smtp', 'login'):
        smtp_login = config.get('smtp', 'login')
    if config.has_option('smtp', 'password'):
       smtp_password = config.get('smtp', 'password')
    smtp = Smtp(smtp_host, smtp_port, smtp_use_tls, smtp_login, smtp_password)
    return smtp


def read_letter(filepath):
    f = open(filepath, 'r')
    subject_line = f.readline()
    subject_line = subject_line.replace('<!--', '').replace('-->', '')
    subject = subject_line.replace('Subject:', '')
    subject = subject.strip()
    content = f.read()
    f.close()
    return subject, content


def send_letter(smtp, email_from, email_to, full_email_path, substitutions={}):
    format = 'plain'
    _, extension = splitext(full_email_path)
    if extension == '.html':
        format = 'html'
    subject, letter = read_letter(full_email_path)

    if substitutions:
        content = letter.format(**substitutions)
    else:
        content = letter

    msg = MIMEText(content, format)
    msg['Subject'] = subject
    msg['From'] = email_from
    msg['To'] = email_to
    smtp.send(email_from, [email_to], msg.as_string())


def send_letter_to_many(smtp, email_from, emails_to, full_email_path, substitutions={}):
    format = 'plain'
    _, extension = splitext(full_email_path)
    if extension == '.html':
        format = 'html'
    subject, letter = read_letter(full_email_path)

    if substitutions:
        content = letter.format(**substitutions)
    else:
        content = letter

    msg = MIMEText(content, format)
    msg['Subject'] = subject
    msg['From'] = email_from
    msg['To'] = ', '.join(emails_to)
    smtp.send(email_from, emails_to, msg.as_string())


class Mail:
    def __init__(self, smtp, from_email, activate_url_template, password_url_template, device_error_email):
        self.device_error_email = device_error_email
        self.smtp = smtp
        self.from_email = from_email
        self.activate_url_template = activate_url_template
        self.password_url_template = password_url_template
        self.path = join(dirname(__file__), '..', 'emails')

    def email_path(self, filename):
        return os.path.join(self.path, filename)

    def send_letter(self, email_to, full_email_path, substitutions={}):
        send_letter(self.smtp, self.from_email, email_to, full_email_path, substitutions)

    def send_activate(self, main_domain, email_to, token):
        url = self.activate_url_template.format(token)
        full_email_path = self.email_path('activate.txt')
        self.send_letter(email_to, full_email_path, dict(main_domain=main_domain, url=url))

    def send_reset_password(self, email_to, token):
        url = self.password_url_template.format(token)
        full_email_path = self.email_path('reset_password.txt')
        self.send_letter(email_to, full_email_path, dict(url=url))

    def send_set_password(self, email_to):
        full_email_path = self.email_path('set_password.txt')
        self.send_letter(email_to, full_email_path)

    def send_logs(self, user_email, data, include_support):
        fd, filename = tempfile.mkstemp()
        with os.fdopen(fd, 'w') as f:
            f.write('Device error report\n')
            f.write('Thank you for sharing Syncloud device error info, Syncloud support will get back to you shortly.\n')
            f.write('If you need to add more details just reply to this email.\n\n')
            f.write(data.encode('utf-8'))
        try:
            from_email = self.device_error_email
            to_email = [user_email]
            if include_support:
                to_email.append(self.device_error_email)
            send_letter_to_many(self.smtp, from_email, to_email, filename)
        finally:
            os.unlink(filename)
