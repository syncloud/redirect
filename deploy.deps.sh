#!/usr/bin/env bash

apt-get update -qq
debconf-set-selections <<< 'mysql-server mysql-server/root_password password root'
debconf-set-selections <<< 'mysql-server mysql-server/root_password_again password root'
apt-get install -y -qq mysql-server libmysqlclient-dev apache2 python python-pip libapache2-mod-wsgi python-mysqldb python-dev openssl
mkdir -p /etc/letsencrypt/live/syncloud.it
openssl req -x509 -newkey rsa:4096 \
  -keyout /etc/letsencrypt/live/syncloud.it/privkey.pem \
  -out /etc/letsencrypt/live/syncloud.it/fullchain.pem \
  -nodes -days 1 \
  -subj "/C=US/ST=Oregon/L=Portland/O=Company Name/OU=Org/CN=www.example.com"