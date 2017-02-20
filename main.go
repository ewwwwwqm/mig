package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"github.com/ewwwwwqm/cli"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"sort"
	"strings"
	"time"
)

// Data section for constants and common phrases.
const (
	AppName    string = "Database migration utility"
	AppVersion string = "0.1.0"

	AppNoInput          string = "No input parameters were specified."
	AppHelpUsage        string = "Use -h to display help information."
	AppAvailableDrivers string = "Available drivers"
	AppHelpURL          string = "Visit github.com/ewwwwwqm/mig for more information."

	AppConnQuery string = "Connection query:"
	AppSQLExec   string = "SQL:"
	AppResult    string = "Result:"
)

// Global varialble for interaction with database.
var db *sql.DB

// Drivers variable for string driver values.
var appDrivers = &availableDrivers{[]string{"sqlite3", "mysql", "postgres"}}

// Type for storing available drivers.
type availableDrivers struct {
	Driver []string
}

// Output method to show available drivers.
func (appDrivers *availableDrivers) Output(list bool) (availableDriversOutput string) {
	sep := ", "
	if list {
		sep = "\n"
	}
	var availableDriversAmount = len(appDrivers.Driver) - 1
	for i, v := range appDrivers.Driver {
		availableDriversOutput += v
		if availableDriversAmount != i {
			availableDriversOutput += sep
		}
	}
	return availableDriversOutput
}

// CheckDriver checks driver from available drivers list.
func CheckDriver(driver string) error {
	for _, v := range appDrivers.Driver {
		if driver == v {
			return nil
		}
	}
	return errors.New("driver: " + driver + " is not available.\n\n")
}

// BuildConn builds connection string.
func BuildConn(args *connT, includeDatabaseName bool) (conn string, err error) {
	if args.Driver == "mysql" {
		includedDbName := args.Dbname
		if !includeDatabaseName {
			includedDbName = ""
		}
		conn = fmt.Sprintf("%v:%v@%v(%v:%v)/%v?charset=%v",
			args.User,
			args.Password,
			args.Protocol,
			args.Host,
			args.Port,
			includedDbName,
			args.Charset)
		return conn, nil
	}
	if args.Driver == "postgres" {
		conn = fmt.Sprintf("host=%s user=%s password=%s sslmode=%v",
			args.Host,
			args.User,
			args.Password,
			args.Sslmode)
		return conn, nil
	}
	if args.Driver == "sqlite3" {
		conn = fmt.Sprintf("%v%v.db", args.Dbpath, args.Dbname)
		return conn, nil
	}
	return "", errors.New("build connection was ignored")
}

// Type for top-level options.
type rootT struct {
	cli.Helper
	Version          bool `cli:"v,version" usage:"display current version"`
	AvailableDrivers bool `cli:"drivers" usage:"display available drivers"`
}

// Root level command.
var root = &cli.Command{
	Desc: AppName + "\nv" + AppVersion + "\n\n" + AppHelpURL,
	Argv: func() interface{} { return new(rootT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*rootT)
		if argv.Version {
			ctx.String(AppVersion + "\n")
		}
		if argv.AvailableDrivers {
			ctx.String(appDrivers.Output(true) + "\n")
		}
		return nil
	},
}

// Type for create command.
type connT struct {
	cli.Helper
	Driver   string `cli:"*drv,driver" usage:"database driver" prompt:"Database driver"`
	Host     string `cli:"host" usage:"hostname or ip" prompt:"Hostname" dft:"127.0.0.1"`
	Protocol string `cli:"protocol" usage:"communication protocol" prompt:"Protocol" dft:"tcp"`
	Port     int16  `cli:"port" usage:"database port" prompt:"Port" dft:"3306"`
	Dbname   string `cli:"*db,dbname" usage:"name of the database" prompt:"Database name"`
	User     string `cli:"u,user" usage:"username" prompt:"Database username"`
	Password string `pw:"p,password" usage:"password" prompt:"Database password"`
	Charset  string `cli:"charset" usage:"character set" prompt:"Database charset" dft:"utf8"`
	Dbpath   string `cli:"dbpath" usage:"database path" prompt:"Database path" dft:"./"`
	Table    string `cli:"tbl,table" usage:"table name" prompt:"Table name" dft:"scheme_info"`
	Sslmode  string `cli:"sslmode" usage:"SSL mode" prompt:"SSL mode" dft:"disable"`
}

// Create command builds connection and tries to create database.
// Returns error if database exists.
var createCom = &cli.Command{
	Name: "create",
	Desc: "Creates database",
	Argv: func() interface{} { return new(connT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*connT)

		if len(argv.Dbname) != 0 {
			ctx.String("\n")
			// pg: postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full

			// check the driver
			err := CheckDriver(argv.Driver)
			if err != nil {
				ctx.String(ctx.Color().Red(err.Error()))
				ctx.String(AppAvailableDrivers + ": " + appDrivers.Output(false) + "\n")
				return nil
			}

			ctx.String(ctx.Color().Bold(AppConnQuery) + "\n")

			// build conn string for specified driver
			conn, err := BuildConn(argv, false)
			if err != nil {
				return fmt.Errorf(err.Error())
			}

			ctx.String(ctx.Color().Dim(conn) + "\n")
			start := time.Now()

			if argv.Driver == "mysql" {
				ctx.String("\n")

				// open connection
				db, err := sql.Open(argv.Driver, conn)
				if err != nil {
					ctx.String(ctx.Color().Red(err.Error()))
					return nil
				}
				defer db.Close()

				ctx.String(ctx.Color().Bold(AppSQLExec) + "\n")

				// create database query
				query := fmt.Sprintf("CREATE DATABASE %s CHARACTER SET %s", argv.Dbname, argv.Charset)
				ctx.String(ctx.Color().Cyan(query+";") + "\n")
				_, err = db.Exec(query)
				if err != nil {
					ctx.String("\n")
					ctx.String(ctx.Color().Red(err.Error()) + "\n")
					return nil
				}

				// use database query
				query = fmt.Sprintf("USE %s", argv.Dbname)
				ctx.String(ctx.Color().Cyan(query+";") + "\n")
				_, err = db.Exec(query)
				if err != nil {
					ctx.String("\n")
					ctx.String(ctx.Color().Red(err.Error()) + "\n")
					return nil
				}
			}
			if argv.Driver == "postgres" {
				ctx.String("\n")

				// open connection
				db, err := sql.Open(argv.Driver, conn)
				if err != nil {
					ctx.String(ctx.Color().Red(err.Error()))
					return nil
				}
				defer db.Close()

				ctx.String(ctx.Color().Bold(AppSQLExec) + "\n")

				// create database query
				query := fmt.Sprintf("CREATE DATABASE %s OWNER %s ENCODING '%s'", 
					argv.Dbname, 
					argv.User, 
					strings.ToUpper(argv.Charset))
				ctx.String(ctx.Color().Cyan(query+";") + "\n")
				_, err = db.Exec(query)
				if err != nil {
					ctx.String("\n")
					ctx.String(ctx.Color().Red(err.Error()) + "\n")
					return nil
				}
			}
			if argv.Driver == "sqlite3" {
				// open connection
				db, err := sql.Open(argv.Driver, conn)
				if err != nil {
					ctx.String(ctx.Color().Red(err.Error()))
					return nil
				}
				defer db.Close()
			}

			elapsed := fmt.Sprintf("%v", time.Since(start))
			ctx.String("\n" +
				ctx.Color().Green("DONE") + " " +
				ctx.Color().Dim("("+elapsed+")") + "\n")
		}

		return nil
	},
}

// Drop command builds connection and tries to drop database.
// Returns error if database exists.
var dropCom = &cli.Command{
	Name: "drop",
	Desc: "Drops database",
	Argv: func() interface{} { return new(connT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*connT)

		if len(argv.Dbname) != 0 {
			ctx.String("\n")

			// check the driver
			err := CheckDriver(argv.Driver)
			if err != nil {
				ctx.String(ctx.Color().Red(err.Error()))
				ctx.String(AppAvailableDrivers + ": " + appDrivers.Output(false) + "\n")
				return nil
			}

			ctx.String(ctx.Color().Bold(AppConnQuery) + "\n")

			// build conn string for specified driver
			conn, err := BuildConn(argv, false)
			if err != nil {
				return fmt.Errorf(err.Error())
			}

			ctx.String(ctx.Color().Dim(conn) + "\n\n")

			// open connection
			db, err := sql.Open(argv.Driver, conn)
			if err != nil {
				ctx.String(ctx.Color().Red(err.Error()))
				return nil
			}
			defer db.Close()

			start := time.Now()

			// drop database query
			if argv.Driver == "mysql" {
				ctx.String(ctx.Color().Bold(AppSQLExec) + "\n")
				query := fmt.Sprintf("DROP DATABASE %s", argv.Dbname)
				ctx.String(ctx.Color().Cyan(query+";") + "\n")
				_, err = db.Exec(query)
				if err != nil {
					ctx.String("\n")
					ctx.String(ctx.Color().Red(err.Error()) + "\n")
					return nil
				}
			}
			if argv.Driver == "postgres" {
				ctx.String(ctx.Color().Bold(AppSQLExec) + "\n")
				query := fmt.Sprintf("DROP DATABASE %s", argv.Dbname)
				ctx.String(ctx.Color().Cyan(query+";") + "\n")
				_, err = db.Exec(query)
				if err != nil {
					ctx.String("\n")
					ctx.String(ctx.Color().Red(err.Error()) + "\n")
					return nil
				}
			}
			if argv.Driver == "sqlite3" {
				// remove the file
				err = os.Remove(conn)
				if err != nil {
					ctx.String("\n")
					ctx.String(ctx.Color().Red(err.Error()) + "\n")
					return nil
				}
			}

			elapsed := fmt.Sprintf("%v", time.Since(start))
			ctx.String("\n" +
				ctx.Color().Green("DONE") + " " +
				ctx.Color().Dim("("+elapsed+")") + "\n")
		}

		return nil
	},
}

// Drop command builds connection and tries to drop database.
// Returns error if database exists.
var describeCom = &cli.Command{
	Name: "describe",
	Desc: "Describes table",
	Argv: func() interface{} { return new(connT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*connT)

		if len(argv.Dbname) != 0 {
			ctx.String("\n")

			// check the driver
			err := CheckDriver(argv.Driver)
			if err != nil {
				ctx.String(ctx.Color().Red(err.Error()))
				ctx.String(AppAvailableDrivers + ": " + appDrivers.Output(false) + "\n")
				return nil
			}

			ctx.String(ctx.Color().Bold(AppConnQuery) + "\n")

			// build conn string for specified driver
			conn, err := BuildConn(argv, true)
			if err != nil {
				return fmt.Errorf(err.Error())
			}

			ctx.String(ctx.Color().Dim(conn) + "\n\n")

			// open connection
			db, err := sql.Open(argv.Driver, conn)
			if err != nil {
				ctx.String(ctx.Color().Red(err.Error()))
				return nil
			}
			defer db.Close()

			ctx.String(ctx.Color().Bold(AppSQLExec) + "\n")
			start := time.Now()

			var ret *sql.Rows
			if argv.Driver == "mysql" || argv.Driver == "sqlite3" {
				// describe table query
				query := fmt.Sprintf("DESCRIBE %v", argv.Table)
				ctx.String(ctx.Color().Cyan(query+";") + "\n")
				ret, err = db.Query(query)
				if err != nil {
					ctx.String("\n")
					ctx.String(ctx.Color().Red(err.Error()) + "\n")
					return nil
				}
			}
			if argv.Driver == "postgres" {
				query := fmt.Sprintf("SELECT attname FROM pg_attribute,pg_class WHERE attrelid=pg_class.oid AND relname='%v' AND attstattarget <> 0;", argv.Table)
				ctx.String(ctx.Color().Cyan(query+";") + "\n")
				ret, err = db.Query(query)
				if err != nil {
					ctx.String("\n")
					ctx.String(ctx.Color().Red(err.Error()) + "\n")
					return nil
				}
			}

			ctx.String("\n")
			ctx.String(ctx.Color().Bold(AppResult) + "\n")

			rrm := RawResultMap(ret)
			rrmLen := len(rrm)
			for k, v := range rrm {
					kk := fmt.Sprintf("%v/%v", k+1, rrmLen)
					ctx.String(ctx.Color().White("{ ") + ctx.Color().Grey(kk) + "\n")
					keys := []string{}
					for ak := range v {
						keys = append(keys, ak)
					}
					sort.Strings(keys)
					for _, av := range keys {
						for akk, avv := range v {
							if akk == av {
								ctx.String("  " + ctx.Color().Grey(akk) + ": " + avv + "\n")
							}
						}
					}

					ctx.String("}")
					if rrmLen == k+1 {
						ctx.String("\n\n")
						rslt := fmt.Sprintf("%v", rrmLen)
						ctx.String("Fetched " + ctx.Color().Bold(rslt) + " result(s).\n")
					} else {
						ctx.String(",\n")
					}
			}

			elapsed := fmt.Sprintf("%v", time.Since(start))
			ctx.String("\n" +
				ctx.Color().Green("DONE") + " " +
				ctx.Color().Dim("("+elapsed+")") + "\n")
		}

		return nil
	},
}

// Drop command builds connection and tries to drop database.
// Returns error if database exists.
var sqlCom = &cli.Command{
	Name: "sql",
	Desc: "Prompts SQL queries",
	Argv: func() interface{} { return new(connT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*connT)

		if len(argv.Dbname) != 0 {
			ctx.String("\n")

			// check the driver
			err := CheckDriver(argv.Driver)
			if err != nil {
				ctx.String(ctx.Color().Red(err.Error()))
				ctx.String(AppAvailableDrivers + ": " + appDrivers.Output(false) + "\n")
				return nil
			}

			ctx.String(ctx.Color().Bold(AppConnQuery) + "\n")

			// build conn string for specified driver
			conn, err := BuildConn(argv, true)
			if err != nil {
				return fmt.Errorf(err.Error())
			}

			ctx.String(ctx.Color().Dim(conn) + "\n\n")

			// open connection
			db, err := sql.Open(argv.Driver, conn)
			if err != nil {
				ctx.String(ctx.Color().Red(err.Error()))
				return nil
			}
			defer db.Close()

			reader := bufio.NewReader(os.Stdin)
			var text string
			for text != "q" {
				ctx.String(ctx.Color().Bold(argv.Driver) + "> ")
				text, _ := reader.ReadString('\n')
				text = strings.TrimSpace(text)

				if text == "q" || text == "exit" || text == "\\q" || text == "/q" || text == ".exit" {
					break
				}

				start := time.Now()

				// get text as query
				query := fmt.Sprintf("%v", text)
				ret, err := db.Query(query)
				if err != nil {
					ctx.String("\n")
					ctx.String(ctx.Color().Red(err.Error()) + "\n")
					continue
				}

				rrm := RawResultMap(ret)
				rrmLen := len(rrm)

				if rrmLen > 0 {
					ctx.String("\n")
				}

				for k, v := range rrm {
					kk := fmt.Sprintf("%v/%v", k+1, rrmLen)
					ctx.String(ctx.Color().White("{ ") + ctx.Color().Grey(kk) + "\n")
					keys := []string{}
					for ak := range v {
						keys = append(keys, ak)
					}
					sort.Strings(keys)
					for _, av := range keys {
						for akk, avv := range v {
							if akk == av {
								ctx.String("  " + ctx.Color().Grey(akk) + ": " + avv + "\n")
							}
						}
					}

					ctx.String("}")
					if rrmLen == k+1 {
						ctx.String("\n\n")
						rslt := fmt.Sprintf("%v", rrmLen)
						ctx.String("Fetched " + ctx.Color().Bold(rslt) + " result(s).\n")
					} else {
						ctx.String(",\n")
					}
				}

				elapsed := fmt.Sprintf("%v", time.Since(start))
				ctx.String("\n" +
					ctx.Color().Green("DONE") + " " +
					ctx.Color().Dim("("+elapsed+")") + "\n")
			}
		}

		return nil
	},
}

// Help variable to store help command.
var help = cli.HelpCommand("Display help information")

// ResultHolder for a query result as an array
// of raw bytes.
func ResultHolder(rows *sql.Rows) []interface{} {
	cols, err := rows.Columns()
	if err != nil {
		return make([]interface{}, 0, 0)
	}
	l := len(cols)
	result := make([]interface{}, l, l)
	for i := range cols {
		result[i] = new(sql.RawBytes)
	}
	return result
}

// RawResult type.
type RawResult map[string]string

// RawResultMap queries results row for each row
// as a map of column to raw bytes.
func RawResultMap(rows *sql.Rows) []RawResult {
	cols, err := rows.Columns()
	if err != nil {
		return make([]RawResult, 0)
	}
	result := ResultHolder(rows)
	results := make([]RawResult, 0)
	c := 0
	for rows.Next() {
		c++
		resultMap := make(map[string]string)
		rows.Scan(result...)
		for i, v := range result {
			f := v.(*sql.RawBytes)
			resultMap[cols[i]] = string((*f)[:])
		}

		results = append(results, resultMap)
	}

	return results
}

func main() {
	if err := cli.Root(root,
		cli.Tree(createCom),
		cli.Tree(describeCom),
		cli.Tree(dropCom),
		cli.Tree(sqlCom),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		if len(os.Args) == 1 {
			fmt.Fprintln(os.Stderr,
				AppName, "\n"+"v"+
					AppVersion, "\n\n"+
					AppHelpUsage, "\n\n"+
					AppAvailableDrivers+": "+
					appDrivers.Output(false)+"\n"+
					AppHelpURL)
			os.Exit(1)
		}
	}
}
