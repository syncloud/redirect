#!/usr/bin/env bash

apt-get update -qq
debconf-set-selections <<< 'mysql-server mysql-server/root_password password root'
debconf-set-selections <<< 'mysql-server mysql-server/root_password_again password root'
apt-get install -y -qq mysql-server libmysqlclient-dev apache2 python python-pip libapache2-mod-wsgi python-mysqldb python-dev