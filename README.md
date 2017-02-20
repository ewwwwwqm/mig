# Database migration utility

[![Godoc Reference][godoc-img]][godoc]
[![Build Status][ci-img]][ci]
[![Go Report Card][goreportcard-img]][goreportcard]

Database migration utility provides a single interface for database management.

Current version: 0.1.0

Included drivers:
* MySQL
* SQLite3
* Postgres

Available commands:
* create
* describe
* drop
* sql

## TODO
- [x] Postgres
- [ ] Create/update/set/delete configuration files for connections
- [ ] Dump database/tables command
- [ ] Pack/unpack dumped database files
- [ ] Merge databases/tables command
- [ ] Webserver api for REST and security

## Usage help

##### Display help information:

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
  describe   Describes table
  drop       Drops database
  sql        Prompts SQL queries
```

##### Display create command help:

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
      --sslmode[=disable]            SSL mode
```

##### Create a database using prompt:

```sh
$ ./mig create
```

Mig prompts you to fill required database parameters:

```sh
Database driver: mysql
Database name: new_db
Database username: root
Database password:
```

After filling required parameters Mig attempts to create a database.

Check the result and operation execution time:

```sql
Connection query:
root:123@tcp(127.0.0.1:3306)/?charset=utf8

SQL:
CREATE DATABASE new_db CHARACTER SET utf8;
USE new_db;

DONE (11.131832ms)
```

If operation fails, Mig outputs database error message:

```sql
Connection query:
root:123@tcp(127.0.0.1:3306)/?charset=utf8

SQL:
CREATE DATABASE new_db CHARACTER SET utf8;

Error 1007: Can't create database 'new_db'; database exists
```

Same example using SQLite3 driver:

```sql
$ ./mig create
Database driver: sqlite3
Database name: new_db
Database username: root
Database password:

Connection query:
./new_db.db

DONE (17.587µs)
```

Same example using Postgres driver:

```sql
$ ./mig create
Database driver: postgres
Database name: new_db
Database username: postgres
Database password:

Connection query:
host=127.0.0.1 user=postgres password=newPassword sslmode=disable

SQL:
CREATE DATABASE new_db OWNER postgres ENCODING 'UTF8';

DONE (335.11233ms)
```

##### Manually create a table and insert data using internal SQL editor:

```sh
$ ./mig sql --driver=mysql -u=root -p=123 --dbname=new_db

Connection query:
root:123@tcp(127.0.0.1:3306)/new_db?charset=utf8
```

After successful connection Mig prompts to enter SQL:

For exit enter <kbd>exit</kbd> or <kbd>\q</kbd> command.

```sql
mysql> CREATE TABLE pet (name VARCHAR(20), owner VARCHAR(20));

DONE (9.864381ms)
mysql> INSERT INTO pet VALUES ("cat", "max");

DONE (11.864013ms)
mysql> SELECT * FROM pet;

{ 1/1
  name: cat
  owner: max
}

Fetched 1 result(s).

DONE (9.593785ms)
mysql> exit
```

##### Describe table command:
```sql
$ ./mig describe --driver=mysql -u=root -p=123 --dbname=new_db --tbl=pet

Connection query:
root:123@tcp(127.0.0.1:3306)/new_db2?charset=utf8

SQL:
DESCRIBE pet;

Result:
{ 1/2
  Default:
  Extra:
  Field: name
  Key:
  Null: YES
  Type: varchar(20)
},
{ 2/2
  Default:
  Extra:
  Field: owner
  Key:
  Null: YES
  Type: varchar(20)
}

Fetched 2 result(s).

DONE (14.334859ms)

```

##### Drop database command:

```sql
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
[goreportcard-img]: https://goreportcard.com/badge/github.com/ewwwwwqm/mig
[goreportcard]: https://goreportcard.com/report/github.com/ewwwwwqm/mig
