#!/usr/bin/env bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

apt-get update -qq
apt-get install -y -qq mysql-client libmysqlclient-dev apache2 python python-pip libapache2-mod-wsgi python-mysqldb python-dev openssl

# integration deps
mkdir -p /etc/letsencrypt/live/syncloud.test
openssl req -x509 -newkey rsa:4096 \
  -keyout /etc/letsencrypt/live/syncloud.test/privkey.pem \
  -out /etc/letsencrypt/live/syncloud.test/fullchain.pem \
  -nodes -days 1 \
  -subj "/C=US/ST=Oregon/L=Portland/O=Company Name/OU=Org/CN=www.example.com"
pip install -r dev_requirements.txt
/etc/letsencrypt/live/syncloud.test/fullchain.pem \
  -nodes -days 1 \
  -subj "/C=US/ST=Oregon/L=Portland/O=Company Name/OU=Org/CN=www.example.com"
pip install -r dev_requirements.txt

mysql --host=mysql --user=root --password=root -e "drop DATABASE redirect"
mysql --host=mysql --user=root --password=root -e "CREATE DATABASE redirect"
mysql --host=mysql --user=root --password=root redirect < ${DIR}/db/init.sql
