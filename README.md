### Deploy

We do not support independent installations of redirect

#### Dependencies

    sudo apt-get install default-mysql-client default-libmysqlclient-dev apache2 libapache2-mod-wsgi openssl confget

#### Get the latest binary

    ./ci/deploy [version]

#### Set credentials (once)

    vim /var/www/redirect/redirect/secret.cfg

#### Configure mysql database (once)

    mysql -ulogin -ppassword -e "create database redirect";

The schema is managed by golang-migrate with migrations embedded in the api
binary (`backend/db/migrations`). The api applies pending migrations on
startup, so an empty database is enough.

### Running locally (development)

#### DB

    sudo docker run --name mysqld -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root -d mysql:5.7.30
    mysql -uroot -proot -hlocalhost --protocol=TCP -e "create database redirect"
    cd backend && go run ./cmd/migrate up --dsn "mysql://root:root@tcp(localhost:3306)/redirect"
    mysql -uroot -proot -hlocalhost --protocol=TCP redirect -e "insert into user (email, password_hash, update_token, active) values ('test@example.com', sha2('password', 256), 'token', 1)"

#### Migrations

    cd backend
    go run ./cmd/migrate up --dsn "mysql://root:root@tcp(localhost:3306)/redirect"
    go run ./cmd/migrate version --dsn "..."
    go run ./cmd/migrate force 19 --dsn "..."

Add a migration as a pair of files in `backend/db/migrations`, numbered from
the highest existing prefix: `0000NN_name.up.sql` and `0000NN_name.down.sql`.

#### StatsD

    sudo docker run --name statsd -p 2003-2004:2003-2004 -p 8125:8125/udp -d graphiteapp/graphite-statsd:1.1.10-4

#### Mail

    sudo docker run --name mail -p 25:25 -d mailhog/mailhog:v1.0.0

#### Web backend

    cd backend
    go build -o www ./cmd/www
    ./www --config-file config/env/local/config.cfg --secret-file config/env/local/secret.cfg --mail-dir emails

#### Web frontend

    npm run dev