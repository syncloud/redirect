[![Stories in Ready](https://badge.waffle.io/syncloud/redirect.png?label=ready&title=Ready)](https://waffle.io/syncloud/redirect)
[![Build Status](https://travis-ci.org/syncloud/redirect.svg?branch=master)](https://travis-ci.org/syncloud/redirect)
### Install dependencies:

    sudo apt-get install apache2 mysql-server python python-pip libapache2-mod-wsgi python-mysqldb git jekyll python-dev libmysqlclient-dev

* set mysql password to root

### Get source code

    sudo useradd redirect
    cd /var/www
    git clone https://github.com/syncloud/redirect.git
    sudo chown -R redirect. redirect

### Install python libs

    cd /var/www/redirect
    sudo pip install -r requirements.txt
    
### Configure apache

    cd /var/www/redirect
    sudo cp apache/redirect.conf /etc/apache2/sites-available/redirect.conf
    sudo a2dissite 000-default.conf
    sudo a2ensite redirect
    
#### Configure apache environment variables

    sudo nano /etc/apache2/envvars
    export SYNCLOUD_DOMAIN=syncloud.it

    sudo service apache2 restart

### Set credentials

    cd /var/www/redirect
    sudo su redirect
    nano redirect/secret.cfg

### Configure mysql database (redirect)

    cd /var/www/redirect
    mysql -uroot -proot -e "create database redirect";
    mysql -ulogin -ppassword < db/init.sql

### Development dependencies
    
    sudo pip install -r dev_requirements.txt

#### Add hosts (local dns)

    sudo sh -c 'echo "127.0.0.1 test.com" >> /etc/hosts'
    sudo sh -c 'echo "127.0.0.1 user.test.com" >> /etc/hosts'

#### Add crontab entry (uat auto deployment)

    crontab -e
    
    */1 * * * * /var/www/redirect/ci/deploy > /var/www/redirect/deploy.log 2>&1

#### Add apache restart to sudoers (uat auto deployment)

    sudo visudo -f /etc/sudoers.d/redirect
    redirect ALL = (root) NOPASSWD: /usr/bin/service apache2 restart
    redirect ALL = (root) NOPASSWD: /usr/bin/pip install -r requirements.txt


#### Upgrade test db from release to master


    sudo su redirect
    cd /var/www/redirect-test
    ./ci/redirectdb redirect redirect-test 006.sql
