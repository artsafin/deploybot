package data

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "gopkg.in/doug-martin/goqu.v3"
    _ "gopkg.in/doug-martin/goqu.v3/adapters/sqlite3"
    "log"
    "os"
)

type dataConn struct {
    *goqu.Database
}

func assertSchemaValid(db *goqu.Database) {
    // select count(*) from sqlite_master where type='table' and tbl_name in ('registrations', 'notifications', 'access_tokens')

    var count int
    found, err := db.From("sqlite_master").Select(goqu.COUNT("*")).Where(goqu.Ex{
            "type": "table",
            "tbl_name": []string{"registrations", "notifications", "access_tokens", "services"},
        }).ScanVal(&count)

    if err != nil || !found || count != 4 {
        log.Fatal("Internal database is corrupt. Do a fresh install with --install flag")
    }
}

func openDb(file string) *goqu.Database {
    db, err := sql.Open("sqlite3", file)
    if err != nil {
        log.Fatal(err)
    }

    return goqu.New("sqlite3", db)
}

func Init(file string) {
    if err := os.Remove(file); err != nil {
        log.Fatal(err)
    }

    sqlStmt := `
    create table notifications (service varchar(100), user_id integer);
    create table registrations (user_id integer not null primary key, chat_id integer, token varchar(100));
    create table access_tokens (token varchar(100) not null primary key, expires integer, used integer);
    create table services (token varchar(100) not null primary key);
    `

    db := openDb(file)
    defer db.Db.Close()

    if _, err := db.Db.Exec(sqlStmt); err != nil {
        log.Fatalf("%q: %s\n", err, sqlStmt)
    }
}

func newDataConn(file string) *dataConn {
    db := openDb(file)

    assertSchemaValid(db)

    return &dataConn{db}
}

