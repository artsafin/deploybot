package main

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
)

type Repository struct {
    db *sql.DB
}

func initScheme(db *sql.DB) {
    sqlStmt := `
    create table if not exists notifications (id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, service varchar(100), user_id integer);
    create table if not exists registrations (id integer not null primary key, chat_id integer, token varchar(100));
    `

    _, err := db.Exec(sqlStmt)
    if err != nil {
        log.Fatalf("%q: %s\n", err, sqlStmt)
    }
}

func NewRepository(file string) *Repository {
    db, err := sql.Open("sqlite3", file)
    if err != nil {
        log.Fatal(err)
    }

    initScheme(db)

    return &Repository{db}
}

func (me *Repository) getNotifications() *Notifications {
    rows, err := me.db.Query("select service, user_id from notifications")
    if err != nil {
        log.Fatal(err)
    }

    defer rows.Close()

    ns := NewNotifications()
    ns.repo = me

    for rows.Next() {
        var service string
        var user_id int
        if err := rows.Scan(&service, &user_id); err == nil {
            LoadNotification(&ns, service, user_id)
        }
    }

    return &ns
}

func (me *Repository) deleteNotification(service string, userId int) {
    me.db.Exec("delete from notifications where service=? and user_id=?", service, userId)
}

func (me *Repository) addNotification(service string, userId int) {
    me.db.Exec("insert into notifications(service, user_id) values(?, ?)", service, userId)
}

