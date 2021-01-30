CREATE TABLE premium_status (
    id integer PRIMARY KEY,
    name varchar(100) NOT NULL
);

insert into premium_status (id, name) values (1, 'inactive');
insert into premium_status (id, name) values (2, 'pending');
insert into premium_status (id, name) values (3, 'active');

ALTER TABLE user ADD column premium_status_id integer NOT NULL DEFAULT 1;
ALTER TABLE user ADD FOREIGN KEY (premium_status_id) REFERENCES premium_status(id);

insert into db_version (version) values ('009');
