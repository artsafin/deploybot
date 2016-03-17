package main

import (
    _ "github.com/mattn/go-sqlite3"
    "gopkg.in/doug-martin/goqu.v3"
    _ "gopkg.in/doug-martin/goqu.v3/adapters/sqlite3"
    "log"
    "fmt"
    "os"
    "time"
    "strconv"
    "crypto/rand"
    "crypto/sha1"
)

func install(config *Config) {
    fmt.Println("Performing initial installation")
    
    initRepository(config.Db.Sqlite)
}

func addToken(config *Config, expires string) string {
    expiresIn, _ := strconv.Atoi(expires)

    rndBytes := make([]byte, 100)

    if _, err := rand.Read(rndBytes); err != nil {
        log.Fatal(err)
    }
    newToken := fmt.Sprintf("%x", sha1.Sum(rndBytes))

    expiresUnix := time.Now().Unix() + int64(expiresIn)

    rec := goqu.Record{"token": newToken, "expires": expiresUnix, "used": 0,}

    _, err := OpenDb(config.Db.Sqlite).From("access_tokens").Insert(rec).Exec()

    if err != nil {
        log.Fatal(err)
    }

    return newToken
}

func initRepository(file string) *goqu.Database {
    if err := os.Remove(file); err != nil {
        log.Fatal(err)
    }

    sqlStmt := `
    create table notifications (service varchar(100), user_id integer);
    create table registrations (user_id integer not null primary key, chat_id integer, token varchar(100));
    create table access_tokens (token varchar(100) not null primary key, expires integer, used integer);
    `

    db := OpenDb(file)

    if _, err := db.Db.Exec(sqlStmt); err != nil {
        log.Fatalf("%q: %s\n", err, sqlStmt)
    }

    return db
}
