#!/bin/bash -e

ENV=$1
SYNCLOUD_DOMAIN=$2

CURRENT=/var/www/redirect/current
DB_VERSION=012

apt install confget
cp -rf ${CURRENT}/config/env/${ENV}/* /var/www/redirect

if  ! id -u redirect > /dev/null 2>&1; then
    adduser --disabled-password --gecos "" redirect
fi
mkdir -p /var/run/redirect
chown redirect. /var/run/redirect
cp ${CURRENT}/config/common/systemd/redirect.api.service /lib/systemd/system/
cp ${CURRENT}/config/common/systemd/redirect.www.service /lib/systemd/system/
systemctl enable redirect.api
systemctl enable redirect.www
systemctl start redirect.api
systemctl start redirect.www

cp ${CURRENT}/config/common/apache/redirect.conf /etc/apache2/sites-available

chown -R redirect. ${CURRENT}/
if a2query -s 000-default; then
  a2dissite 000-default
fi
if ! a2query -s redirect; then
  a2ensite redirect
fi
a2enmod rewrite
a2enmod ssl
a2enmod proxy
a2enmod proxy_http
echo "export SYNCLOUD_DOMAIN=${SYNCLOUD_DOMAIN}" >> /etc/apache2/envvars
grep SYNCLOUD_DOMAIN /etc/apache2/envvars

crontab -u redirect ${CURRENT}/config/common/cron/crontab
crontab -u redirect -l

DB_HOST=$(confget -f /var/www/redirect/config.cfg -s mysql host)
if ! mysql --host=${DB_HOST} --user=root --password=root -e 'use redirect'; then
  echo "init redirect database"
  mysql --host=${DB_HOST} --user=root --password=root -e "CREATE DATABASE redirect"
  mysql --host=${DB_HOST} --user=root --password=root redirect < ${CURRENT}/db/init.sql
fi

if mysql --host=${DB_HOST} --user=root --password=root redirect -e 'select version from db_version' | grep $DB_VERSION; then
  echo "database is up to date"
else
  echo "updating redirect database to $DB_VERSION"
  mysql --host=${DB_HOST} --user=root --password=root redirect < ${CURRENT}/db/update.sql
fi

service apache2 restart
