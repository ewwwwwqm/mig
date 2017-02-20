package main

import (
	"testing"
)

func TestCheckDriver(t *testing.T) {
	// Validate driver exceptions
	if CheckDriver("mysql") != nil {
		t.Error("expected nil")
	}
	if CheckDriver("sqlite3") != nil {
		t.Error("expected nil")
	}
	if CheckDriver("postgres") != nil {
		t.Error("expected nil")
	}
	if CheckDriver("invalid_driver") == nil {
		t.Error("invalid driver")
	}
}

func TestBuildConn(t *testing.T) {
	var args connT // input parameters

	// Set arguments
	args.Driver = "mysql"
	args.Dbname = "db_new"
	args.User = "root"
	args.Password = "123"
	args.Host = "127.0.0.1"
	args.Port = 3306
	args.Charset = "utf8"

	// Try to build connection string
	conn, err := BuildConn(&args, true)
	if err != nil {
		t.Error("failed to build connection")
	}
	if conn != "root:123@(127.0.0.1:3306)/db_new?charset=utf8" {
		t.Error("invalid connection")
	}
}
