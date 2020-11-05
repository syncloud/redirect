#!/bin/bash -e

ENV=$1
SYNCLOUD_DOMAIN=$2

CURRENT=/var/www/redirect/current

pip install -r ${CURRENT}/requirements.txt

cp -rf ${CURRENT}/config/env/${ENV}/* /var/www/redirect

cp ${CURRENT}/config/common/systemd/redirect.service /lib/systemd/system/
systemctl enable redirect
systemctl start redirect

cp ${CURRENT}/config/common/apache/redirect.conf /etc/apache2/sites-available

if  ! id -u redirect > /dev/null 2>&1; then
    adduser --disabled-password --gecos "" redirect
fi
chown -R redirect. ${CURRENT}
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
crontab ${CURRENT}/config/common/cron/crontab
crontab -l
service apache2 restart