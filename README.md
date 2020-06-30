### Deploy

We do not support independent installs of redirect

#### Dependencies

    sudo apt-get install apache2 mysql-server python python-pip libapache2-mod-wsgi python-mysqldb python-dev libmysqlclient-dev

#### Get the latest binary

    ./ci/deploy [version]

### Set credentials (once)

    vim /var/www/redirect/redirect/secret.cfg

### Configure mysql database (once)

    mysql -ulogin -ppassword -e "create database redirect";
    mysql -ulogin -ppassword < db/init.sql

