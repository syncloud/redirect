#!/bin/bash -e

sed -i '/SYNCLOUD_DOMAIN.*/d' /etc/apache2/envvars
a2dissite redirect
service apache2 restart

systemctl stop redirect.api
systemctl stop redirect.www
systemctl disable redirect.api
systemctl disable redirect.www
crontab -u redirect -r || true
rm /lib/systemd/system/redirect.api.service
rm /lib/systemd/system/redirect.www.service
