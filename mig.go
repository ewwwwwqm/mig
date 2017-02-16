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
	APP_NAME string 			= "Database migration utility"
	APP_VERSION string 			= "0.1.0"

	APP_NO_INPUT string 		= "No input parameters were specified."
	APP_HELP_USAGE string 		= "Use -h to display help information."
	APP_AVAIL_DRIVERS string 	= "Available drivers"
	APP_HELP_URL string 		= "Visit github.com/ewwwwwqm/mig for more information."

	APP_CONN_QUERY string 		= "Connection query:"
	APP_SQL_EXEC string 		= "SQL:"
	APP_RESULT string 			= "Result:"
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

// Checks driver from available drivers list.
func checkDriver(driver string) error {
	for _, v := range appDrivers.Driver {
		if driver == v {
			return nil
		}
	}
	return errors.New("driver: " + driver + " is not available.\n\n")
}

// Builds connection string.
func buildConn(args *connT, includeDatabaseName bool) (err error, conn string) {
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
		return nil, conn
	}
	if args.Driver == "sqlite3" {
		conn = fmt.Sprintf("%v%v.db", args.Dbpath, args.Dbname)
		return nil, conn
	}
	return errors.New("build connection was ignored"), ""
}

// Type for top-level options.
type rootT struct {
	cli.Helper
	Version  bool `cli:"v,version" usage:"display current version"`
	AvailableDrivers  bool `cli:"drivers" usage:"display available drivers"`
}

// Root level command.
var root = &cli.Command{
	Desc: APP_NAME + "\nv" + APP_VERSION + "\n\n" + APP_HELP_URL,
	Argv: func() interface{} { return new(rootT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*rootT)
		if argv.Version {
			ctx.String(APP_VERSION + "\n")
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
	Driver      string        `cli:"*drv,driver" usage:"database driver" prompt:"Database driver"`
	Host        string        `cli:"host" usage:"hostname or ip" prompt:"Hostname" dft:"127.0.0.1"`
	Protocol 	string 		  `cli:"protocol" usage:"communication protocol" prompt:"Protocol" dft:"tcp"`
	Port        int16         `cli:"port" usage:"database port" prompt:"Port" dft:"3306"`
	Dbname      string        `cli:"*db,dbname" usage:"name of the database" prompt:"Database name"`
	User        string        `cli:"u,user" usage:"username" prompt:"Database username"`
	Password    string        `pw:"p,password" usage:"password" prompt:"Database password"`
	Charset     string        `cli:"charset" usage:"character set" prompt:"Database charset" dft:"utf8"`
	Dbpath		string 		  `cli:"dbpath" usage:"database path" prompt:"Database path" dft:"./"`
	Table 		string 		  `cli:"tbl,table" usage:"table name" prompt:"Table name" dft:"scheme_info"`
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
			err := checkDriver(argv.Driver)
			if err != nil {
				ctx.String(ctx.Color().Red(err.Error()))
				ctx.String(APP_AVAIL_DRIVERS + ": " + appDrivers.Output(false) + "\n")
				return nil
			}
			
			ctx.String(ctx.Color().Bold(APP_CONN_QUERY) + "\n")

			// build conn string for specified driver
			err, conn := buildConn(argv, false)
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

				ctx.String(ctx.Color().Bold(APP_SQL_EXEC) + "\n")		

				// create database query
				query := fmt.Sprintf("CREATE DATABASE %s CHARACTER SET %s", argv.Dbname, argv.Charset)
				ctx.String(ctx.Color().Cyan(query + ";") + "\n")
				_, err = db.Exec(query)
				if err != nil {
					ctx.String("\n")
					ctx.String(ctx.Color().Red(err.Error()) + "\n")
					return nil
				}

				// use database query
				query = fmt.Sprintf("USE %s", argv.Dbname)
				ctx.String(ctx.Color().Cyan(query + ";") + "\n")
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
				ctx.Color().Dim("(" + elapsed + ")") + "\n")
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
			err := checkDriver(argv.Driver)
			if err != nil {
				ctx.String(ctx.Color().Red(err.Error()))
				ctx.String(APP_AVAIL_DRIVERS + ": " + appDrivers.Output(false) + "\n")
				return nil
			}
			
			ctx.String(ctx.Color().Bold(APP_CONN_QUERY) + "\n")

			// build conn string for specified driver
			err, conn := buildConn(argv, false)
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
				ctx.String(ctx.Color().Bold(APP_SQL_EXEC) + "\n")
				query := fmt.Sprintf("DROP DATABASE %s", argv.Dbname)
				ctx.String(ctx.Color().Cyan(query + ";") + "\n")
				_, err = db.Exec(query)
				if err != nil {
					ctx.String("\n")
					ctx.String(ctx.Color().Red(err.Error()) + "\n")
					return nil
				}
			}
			// remove the file
			if argv.Driver == "sqlite3" {
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
				ctx.Color().Dim("(" + elapsed + ")") + "\n")
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
			err := checkDriver(argv.Driver)
			if err != nil {
				ctx.String(ctx.Color().Red(err.Error()))
				ctx.String(APP_AVAIL_DRIVERS + ": " + appDrivers.Output(false) + "\n")
				return nil
			}
			
			ctx.String(ctx.Color().Bold(APP_CONN_QUERY) + "\n")

			// build conn string for specified driver
			err, conn := buildConn(argv, true)
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

			ctx.String(ctx.Color().Bold(APP_SQL_EXEC) + "\n")
			start := time.Now()

			// describe table query
			query := fmt.Sprintf("DESCRIBE %v", argv.Table)
			ctx.String(ctx.Color().Cyan(query + ";") + "\n")
			ret, err := db.Query(query)		
			if err != nil {
				ctx.String("\n")
				ctx.String(ctx.Color().Red(err.Error()) + "\n")
				return nil
			}

			ctx.String("\n")
			ctx.String(ctx.Color().Bold(APP_RESULT) + "\n")

			rrm := RawResultMap(ret)
			rrmLen := len(rrm)
			for k,v := range rrm {
				kk := fmt.Sprintf("%v/%v", k+1, rrmLen)
				ctx.String(ctx.Color().Bold("# " + kk) + "\n")
				keys := []string{}
				for ak,_ := range v {
					keys = append(keys, ak)
				}
				sort.Strings(keys)
				for _,av := range keys {
					for akk,avv := range v {
						if akk == av {
							ctx.String("\t" + ctx.Color().Grey(akk) + ": " + avv + "\n")
						}
					}
				}
				ctx.String("\n")
				if rrmLen == k+1 {
					rslt := fmt.Sprintf("%v", rrmLen)
					ctx.String("Fetched " + ctx.Color().Bold(rslt) + " result(s).\n")
				}
			}

			elapsed := fmt.Sprintf("%v", time.Since(start))
			ctx.String("\n" +
				ctx.Color().Green("DONE") + " " + 
				ctx.Color().Dim("(" + elapsed + ")") + "\n")
		}

		return nil
	},
}

// Drop command builds connection and tries to drop database.
// Returns error if database exists.
var sqlCom = &cli.Command{
	Name: "sql",
	Desc: "Promts for SQL queries",
	Argv: func() interface{} { return new(connT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*connT)

		if len(argv.Dbname) != 0 {
			ctx.String("\n")

			// check the driver
			err := checkDriver(argv.Driver)
			if err != nil {
				ctx.String(ctx.Color().Red(err.Error()))
				ctx.String(APP_AVAIL_DRIVERS + ": " + appDrivers.Output(false) + "\n")
				return nil
			}
			
			ctx.String(ctx.Color().Bold(APP_CONN_QUERY) + "\n")

			// build conn string for specified driver
			err, conn := buildConn(argv, true)
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
				ctx.String(ctx.Color().Bold("SQL") + " > ")
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

				for k,v := range rrm {
					kk := fmt.Sprintf("%v/%v", k+1, rrmLen)
					ctx.String(ctx.Color().Bold("# " + kk) + "\n")
					keys := []string{}
					for ak,_ := range v {
						keys = append(keys, ak)
					}
					sort.Strings(keys)
					for _,av := range keys {
						for akk,avv := range v {
							if akk == av {
								ctx.String("\t" + ctx.Color().Grey(akk) + ": " + avv + "\n")
							}
						}
					}

					ctx.String("\n")
					if rrmLen == k+1 {
						rslt := fmt.Sprintf("%v", rrmLen)
						ctx.String("Fetched " + ctx.Color().Bold(rslt) + " result(s).\n")
					}
				}

				elapsed := fmt.Sprintf("%v", time.Since(start))
				ctx.String("\n" +
					ctx.Color().Green("DONE") + " " + 
					ctx.Color().Dim("(" + elapsed + ")") + "\n")
			}
		}

		return nil
	},
}

// Help variable to store help command.
var help = cli.HelpCommand("Display help information")

// Creates a result holder for a query result
// as an array of raw bytes.
func ResultHolder(rows *sql.Rows) []interface{} {
	cols, err := rows.Columns()
	if err != nil {
		return make([]interface{}, 0, 0)
	}	
	l := len(cols)
	result := make([]interface{}, l, l)
	for i, _ := range cols {
		result[i] = new(sql.RawBytes)
	}
	return result
}

// Type for raw result.
type rawResult map[string]string

// Converts a query results row for each row result
// as a map of column to raw bytes.
func RawResultMap(rows *sql.Rows) []rawResult {
	cols, err := rows.Columns()
	if err != nil {
		return make([]rawResult, 0)
	}
	result := ResultHolder(rows)
	results := make([]rawResult, 0)
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
		cli.Tree(dropCom),
		cli.Tree(describeCom),
		cli.Tree(sqlCom),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} else {
		if len(os.Args) == 1 {
			fmt.Fprintln(os.Stderr,
				APP_NAME, "\n" + "v" +
				APP_VERSION, "\n\n" +
				APP_HELP_USAGE, "\n\n" +
				APP_AVAIL_DRIVERS + ": " + 
				appDrivers.Output(false) + "\n" +
				APP_HELP_URL)
			os.Exit(1)
		}
	}
}
