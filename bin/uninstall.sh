#!/bin/bash -e

sed -i '/SYNCLOUD_DOMAIN.*/d' /etc/apache2/envvars
a2dissite redirect
service apache2 restart

systemctl stop redirect
systemctl disable redirect
rm /lib/systemd/system/redirect.service
