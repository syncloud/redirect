#!/bin/bash -e
apt-get update -qq
debconf-set-selections <<< "postfix postfix/mailname string your.hostname.com"
debconf-set-selections <<< "postfix postfix/main_mailer_type string 'Internet Site'"
apt-get install -y -qq mysql-client libmysqlclient-dev postfix
pip install -r requirements.txt
pip install -r dev_requirements.txt
adduser --disabled-password --gecos "" test
mkdir mail.root
chown test. mail.root
