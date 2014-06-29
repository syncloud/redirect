create table db_version (
    version varchar(10) not null,
    last_update TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

insert into db_version (version) values ('001');