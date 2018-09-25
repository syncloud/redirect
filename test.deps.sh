#!/usr/bin/env bash

apt-get update -qq
debconf-set-selections <<< "postfix postfix/mailname string your.hostname.com"
debconf-set-selections <<< "postfix postfix/main_mailer_type string 'Internet Site'"
apt-get install -y -qq mysql-server postfix
pip install -r requirements.txt
pip install -r dev_requirements.txt
