package cli

// TODO: port notification

/*
#!/usr/bin/env python
import sys
from os.path import dirname, join, normpath, exists
sys.path.append(normpath(join(dirname(__file__), '..')))

import os
import argparse
import traceback
from datetime import datetime
import time
from redirect import db_helper, mail, config


def send(smtp, email_from, email_to, filepath):
    try:
        print("Sending email to: {}".format(email_to))
        mail.send_letter(smtp, email_from, email_to, filepath)
    except Exception as e:
        print("Failed to send letter to: {}".format(email_to))
        traceback.print_exc()


def items_per_second(limit, items, f):
    counter = 0
    start = datetime.now()
    for item in items:
        if counter == limit:
            elapsed = (datetime.now() - start).total_seconds()
            if elapsed < 1:
                time.sleep(1 - elapsed)
            start = datetime.now()
            counter = 0
        f(item)
        counter += 1

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='This is redirect email sending tool', usage='%(prog)s [options]')
    parser.add_argument('filepath', help='path to file with email content')
    parser.add_argument('email', help='recipient, specify ALL to send to all users from database, specify SQL script filename to send to specific users')
    parser.add_argument('--from', dest='from_email', help='email to send from, will be taken from redirect config if not provided')
    args = parser.parse_args()

    filepath = args.filepath
    email_from = args.from_email

    email_to = args.email

    try:
        if not os.path.exists(filepath):
            raise Exception("Can't find letter file {}".format(filepath))

        the_config = config.read_redirect_configs()
        if not email_from:
            email_from = the_config.get('mail', 'support')

        create_storage = db_helper.get_storage_creator(the_config)
        smtp = mail.get_smtp(the_config)

        if email_to == 'ALL' or email_to.endswith('.sql'):
            if email_to == 'ALL':
                with create_storage() as storage:
                    items_per_second(5, storage.users_iterate(), lambda user: send(smtp, email_from, user.email, filepath))
            if email_to.endswith('.sql'):
                emails_script = email_to
                if not exists(emails_script):
                    raise Exception("Can't find emails query file: {}".format(emails_script))
                with create_storage() as storage:
                    with open(emails_script, 'r') as f:
                        query = f.read()
                        print 'Running script: {}'.format(query)
                        emails = storage.get_users_emails(query)
                        users = [storage.get_user_by_email(email) for email in emails]
                        items_per_second(5, users, lambda user: send(smtp, email_from, user.email, filepath))
        else:
            send(smtp, email_from, email_to, filepath)

    except Exception as e:
        print("Error happened while sending emails")
        traceback.print_exc()
        exit(1)


*/
