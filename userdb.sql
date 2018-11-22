CREATE DATABASE userdb_test;
\c userdb_test
CREATE TABLE users (
	id serial primary key,
	ih char(64) not null,
	verifier char(585) not null,
	username varchar(64) not null,
	email varchar(128) default null,
	totpKey char(64) default null
);
