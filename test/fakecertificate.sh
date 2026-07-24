#!/bin/bash -e

mkdir -p /etc/letsencrypt/live/syncloud.test
openssl req -x509 -newkey rsa:4096 \
  -keyout /etc/letsencrypt/live/syncloud.test/privkey.pem \
  -out /etc/letsencrypt/live/syncloud.test/fullchain.pem \
  -nodes -days 1 \
  -subj "/C=US/ST=Oregon/L=Portland/O=Company Name/OU=Org/CN=www.example.com"

