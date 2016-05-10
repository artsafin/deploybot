package data

import (
    "sync"
    "log"
    "fmt"
    "gopkg.in/doug-martin/goqu.v3"
)

func newNotifications(db *dataConn) *Notifications {
    return &Notifications{db: db}
}

type Notifications struct {
    db *dataConn
    lock sync.RWMutex
}

type Notification struct {
    Service string `db:"service"`
    UserId  int `db:"user_id"`
}

func (me *Notifications) getServiceUserIds(service string) []int {
    userIds := make([]int, 0)

    if err := me.db.From("notifications").Select("user_id").Where(goqu.Ex{"service": service}).ScanVals(userIds); err != nil {
        log.Fatal(err)
    }

    return userIds
}

func (me *Notifications) has(service string, userId int) bool {
    rec := goqu.Ex{"service": service, "user_id": userId}

    count, err := me.db.From("notifications").Where(rec).Count()

    if err != nil {
        log.Fatal(err)
    }

    return count > 0
}

func (me *Notifications) Delete(service string, userId int) error {
    me.lock.Lock()
    defer me.lock.Unlock()

    if !me.has(service, userId) {
        return fmt.Errorf("Cannot delete")
    }

    rec := goqu.Ex{"service": service, "user_id": userId}

    if _, err := me.db.From("notifications").Where(rec).Delete().Exec(); err != nil {
        log.Fatal(err)
    }

    return nil
}

func (me *Notifications) Add(service string, userId int) error {
    me.lock.Lock()
    defer me.lock.Unlock()

    if me.has(service, userId) {
        return fmt.Errorf("Already added")
    }

    rec := goqu.Record{"service": service, "user_id": userId}

    if _, err := me.db.From("notifications").Insert(rec).Exec(); err != nil {
        log.Fatal(err)
    }

    return nil
}