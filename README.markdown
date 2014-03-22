### Install dependencies:

    sudo apt-get install apache2 mysql-server python python-pip libapache2-mod-wsgi python-mysqldb
    sudo pip install -r requirements.txt

### Configure apache:

    sudo cp apache/redirect.conf /etc/apache2/sites-available/redirect.conf
    sudo a2dissite 000-default.conf
    sudo a2ensite redirect
    sudo useradd -m redirect
    sudo service apache2 restart

### Configure mail server

    sudo apt-get install exim4
    sudo dpkg-reconfigure exim4-config

* internet site: Y
* leave the rest as is and never allow relay

### Configure mysql database (redirect)

    mysql -u login -p password < db/init.sql

### Development dependencies
    
    sudo pip install -r dev_requirements.txt

### Configuration

Copy redirect/config.cfg.dist to redirect/config.cfg
and set all needed config properties


### Run tests

    py.test


[![Build Status](https://travis-ci.org/syncloud/redirect.svg?branch=master)](https://travis-ci.org/syncloud/redirect)
