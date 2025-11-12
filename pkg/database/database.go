package database

import (
  "errors"
  "database/sql"
  "github.com/go-sql-driver/mysql"
)

var db* sql.DB = nil;
func Prepare() error {
  if db == nil {
    return errors.New("Must successfully call database.Open() before calling database.Prepare()");
  }


  if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
    id BINARY(16) UNIQUE NOT NULL PRIMARY KEY,
    email VARCHAR(255) UNIQUE
    session_id VARCHAR(128)
    session_expiry TIMESTAMP
  )`); err != nil { return err; }

  if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS notes (
    id BINARY(16) UNIQUE NOT NULL PRIMARY KEY,
    title TEXT,
    owner BINARY(16) NOT NULL,
    FOREIGN KEY(owner) REFERENCES users(id)
  )`); err != nil { return err; }

  return nil;
}

func Open(user string, pass string, db_name string) (*sql.DB, error) {
  db_cfg := mysql.NewConfig();
  db_cfg.User = user;
  db_cfg.Passwd = pass;
  db_cfg.Net = "tcp";
  db_cfg.Addr = "127.0.0.1:3306";
  db_cfg.DBName = db_name;
  db_cfg.Params = map[string]string {"parseTime": "true"};

  var err error;
  db, err = sql.Open("mysql", db_cfg.FormatDSN());
  if err != nil {
    return nil, err;
  }

  err = db.Ping();
  if err != nil {
    return nil, err;
  }

  return db, nil;
}
