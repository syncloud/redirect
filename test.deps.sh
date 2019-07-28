#!/usr/bin/env bash

apt-get update -qq
debconf-set-selections <<< "postfix postfix/mailname string your.hostname.com"
debconf-set-selections <<< "postfix postfix/main_mailer_type string 'Internet Site'"
debconf-set-selections <<< 'mysql-server mysql-server/root_password password root'
debconf-set-selections <<< 'mysql-server mysql-server/root_password_again password root'
apt-get install -y -qq mysql-server libmysqlclient-dev postfix
pip install -r requirements.txt
pip install -r dev_requirements.txt
adduser --disabled-password --gecos "" test
mkdir mail.root
chown test. mail.root