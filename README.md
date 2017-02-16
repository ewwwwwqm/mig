# Database migration utility

[![Godoc Reference][godoc-img]][godoc]
[![Build Status][ci-img]][ci]
[![Go Report Card][goreportcard-img]][goreportcard]

Database migration utility provides a single interface for database management.

Current version: 0.1.0

Included drivers:
* mysql
* sqlite3
* postgres *(currently not available)*

Available commands:
* create
* drop
* describe
* sql

## TODO
* Postgres
* Create/update/set/delete configuration files for connections
* Dump database/tables command
* Pack/unpack dumped database files
* Merge databases/tables command
* Webserver api for REST and security

## Examples

Display help information:

```sh
$ ./mig -h

Database migration utility
v0.1.0

Visit github.com/ewwwwwqm/mig for more information.

Options:

  -h, --help      display help information
  -v, --version   display current version
      --drivers   display available drivers

Commands:

  create     Creates database
  drop       Drops database
  describe   Describes table
  sql        Promts for SQL queries
```

Create command help example:

```sh
$ ./mig create -h

Creates database

Options:

  -h, --help                         display help information
      --drv, --driver               *database driver
      --host[=127.0.0.1]             hostname or ip
      --protocol[=tcp]               communication protocol
      --port[=3306]                  database port
      --db, --dbname                *name of the database
  -u, --user                         username
  -p, --password                     password
      --charset[=utf8]               character set
      --dbpath[=./]                  database path
      --tbl, --table[=scheme_info]   table name
```

Create a database using prompt:

```sh
$ ./mig create

Database driver: mysql
Database name: new_db
Database username: root
Database password:

Connection query:
root:123@tcp(127.0.0.1:3306)/?charset=utf8

SQL:
CREATE DATABASE new_db CHARACTER SET utf8;
USE new_db;

DONE (11.131832ms)
```

Create a table and insert data using sql command:

```sh
$ ./mig sql --driver=mysql -u=root -p=123 --dbname=new_db

Connection query:
root:123@tcp(127.0.0.1:3306)/new_db?charset=utf8

SQL > CREATE TABLE pet (name VARCHAR(20), owner VARCHAR(20));

DONE (9.864381ms)
SQL > INSERT INTO pet VALUES ("cat", "max");

DONE (11.864013ms)
SQL > SELECT * FROM pet;

# 1/1
	name: cat
	owner: max

Fetched 1 result(s).

DONE (9.593785ms)
SQL > exit
```

Describe table command:
```sh
$ ./mig describe --driver=mysql -u=root -p=123 --dbname=new_db --tbl=pet

Connection query:
root:123@tcp(127.0.0.1:3306)/new_db?charset=utf8

SQL:
DESCRIBE pet;

Result:
# 1/2
	Default:
	Extra:
	Field: name
	Key:
	Null: YES
	Type: varchar(20)

# 2/2
	Default:
	Extra:
	Field: owner
	Key:
	Null: YES
	Type: varchar(20)

Fetched 2 result(s).

DONE (636.755µs)
```

Drop database command:

```sh
$ ./mig drop --driver=mysql -u=root -p=123 --dbname=new_db

Connection query:
root:123@tcp(127.0.0.1:3306)/?charset=utf8

SQL:
DROP DATABASE new_db;

DONE (12.676195ms)
```

## Thanks
* https://github.com/mkideal/cli package for building CLI apps.
* https://github.com/labstack/gommon/tree/master/color for colors.
* https://github.com/go-sql-driver/mysql MySQL driver.
* https://github.com/lib/pq Postgres driver.
* https://github.com/mattn/go-sqlite3 SQLite3 driver.
* https://github.com/bennAH/sqlutil/blob/master/rows.go raw result map.

[godoc]: http://godoc.org/github.com/ewwwwwqm/mig
[godoc-img]: https://godoc.org/github.com/ewwwwwqm/mig?status.svg
[ci-img]: https://travis-ci.org/ewwwwwqm/mig.svg?branch=master
[cov-img]: https://coveralls.io/repos/github/ewwwwwqm/migbadge.svg?branch=master
[ci]: https://travis-ci.org/ewwwwwqm/mig
[cov]: https://coveralls.io/github/ewwwwwqm/mig?branch=master
[goreportcard-img]: https://goreportcard.com/badge/github.com/ewwwwwqm/mig?etag=1
[goreportcard]: https://goreportcard.com/report/github.com/ewwwwwqm/mig
