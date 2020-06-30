#!/bin/bash -e
apt-get update -qq
apt-get install -y -qq mysql-client libmysqlclient-dev
pip install -r requirements.txt
pip install -r dev_requirements.txt
adduser --disabled-password --gecos "" test
mkdir mail.root
chown test. mail.root
./ci/recreatedb
