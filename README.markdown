[![Stories in Ready](https://badge.waffle.io/syncloud/redirect.png?label=ready&title=Ready)](https://waffle.io/syncloud/redirect)
[![Build Status](https://travis-ci.org/syncloud/redirect.svg?branch=master)](https://travis-ci.org/syncloud/redirect)
### Install dependencies:

    sudo apt-get install apache2 mysql-server python python-pip libapache2-mod-wsgi python-mysqldb git
    sudo pip install -r requirements.txt

* set mysql password to root

### Get source code

    sudo useradd redirect
    cd /var/www
    git clone https://github.com/syncloud/redirect.git
    sudo chown -R redirect. redirect
    
### Configure apache

    cd /var/www/redirect
    sudo cp apache/redirect.conf /etc/apache2/sites-available/redirect.conf
    sudo a2dissite 000-default.conf
    sudo a2ensite redirect
    
#### Configure apache environment variables

    sudo nano /etc/apache2/envvars
    export SYNCLOUD_DOMAIN=syncloud.it
    export SYNCLOUD_ENV=prod
    export SYNCLOUD_PORT=80
    
    sudo service apache2 restart

### Configure mail server

    sudo apt-get install exim4
    sudo dpkg-reconfigure exim4-config

* internet site: Y
* leave the rest as is and never allow relay

### Configure mysql database (redirect)

    cd /var/www/redirect
    mysql -u root -p root -e "create database redirect";
    mysql -u login -p password < db/init.sql

### Development dependencies
    
    sudo pip install -r dev_requirements.txt

### Configuration

Copy redirect/config.cfg.dist to redirect/config.cfg
and set all needed config properties


### Run tests

    py.test

#### Add hosts (local dns)

    sudo sh -c 'echo "127.0.0.1 test.com" >> /etc/hosts'
    sudo sh -c 'echo "127.0.0.1 user.test.com" >> /etc/hosts'

#### Create and edit config

    cp config.cfg.dist config.cfg

#### Setup apache site (and set WSGIScriptAlias path)

    sudo cp apache/redirect.conf /etc/apache2/sites-available
    sudo a2ensite redirect
    sudo service apache2 restart

#### Add crontab entry (auto deployment)

    crontab -e
    
    */1 * * * * /var/www/redirect-test/deploy.sh > /var/www/redirect-test/deploy.log

#### Add apache restart to sudoers (auto deployment)

    sudo visudo -f /etc/sudoers.d/redirect
    redirect ALL = (root) NOPASSWD: /usr/bin/service apache2-test restart
    redirect ALL = (root) NOPASSWD: /usr/bin/pip install -r requirements.txt


#### Upgrade test db from release to master


    sudo su redirect
    cd /var/www/redirect-test
    ./ci/redirectdb redirect redirect-test 006.sql
