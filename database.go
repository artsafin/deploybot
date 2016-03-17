package main

import (
    "database/sql"
    "gopkg.in/doug-martin/goqu.v3"
    _ "gopkg.in/doug-martin/goqu.v3/adapters/sqlite3"
    "log"
)

type Repository struct {
    *goqu.Database
}

func assertSchemeValid(db *goqu.Database) {
    // select count(*) from sqlite_master where type='table' and tbl_name in ('registrations', 'notifications', 'access_tokens')

    var count int
    found, err := db.From("sqlite_master").Select(goqu.COUNT("*")).Where(goqu.Ex{
            "type": "table",
            "tbl_name": []string{"registrations", "notifications", "access_tokens"},
        }).ScanVal(&count)

    if err != nil || !found || count != 3 {
        log.Fatal("Internal database is corrupt. Do a fresh install with --install flag")
    }
}

func OpenDb(file string) *goqu.Database {
    db, err := sql.Open("sqlite3", file)
    if err != nil {
        log.Fatal(err)
    }

    return goqu.New("sqlite3", db)
}

func NewRepository(file string) *Repository {
    db := OpenDb(file)

    assertSchemeValid(db)

    return &Repository{db}
}

