#!/usr/bin/env bash
apt-get install -y libmysqlclient-dev
apt-get install -y python-dev
apt-get install -y -qq postfix

pip install -r requirements.txt
pip install -r dev_requirements.txt

apt-get install -y ruby ruby-dev make gcc nodejs
gem install jekyll --no-rdoc --no-ri

