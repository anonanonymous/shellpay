-- gateway service database
CREATE DATABASE signatures;
\c signatures;
CREATE TABLE keys (
	id SERIAL PRIMARY KEY,
	pub CHAR(64) NOT NULL,
	priv CHAR(128) NOT NULL,
	address VARCHAR(187) NOT NULL,
	expiry INT NOT NULL
);
