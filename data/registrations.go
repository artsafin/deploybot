package data

import (
    "sync"
    "log"
    "gopkg.in/doug-martin/goqu.v3"
)

func newRegistrations(db *dataConn) *Registrations {
    return &Registrations{db: db}
}

type Registration struct {
    UserId int `db:"user_id"`
    ChatId int64 `db:"chat_id"`
    Token string `db:"token"`
}

type Registrations struct {
    db *dataConn
    lock sync.RWMutex
}

func (me *Registrations) Has(userId int) bool {
    me.lock.RLock()
    defer me.lock.RUnlock()

    conds := goqu.Ex{"user_id": userId}

    count, err := me.db.From("registrations").Select("chat_id", "user_id", "token").Where(conds).Count()

    if err != nil {
        return false
    }

    return count > 0
}

func (me *Registrations) get(userId int, chatId int64) (*Registration, bool) {
    me.lock.RLock()
    defer me.lock.RUnlock()

    u := Registration{}

    conds := goqu.Ex{"user_id": userId, "chat_id": chatId}

    if found, err := me.db.From("registrations").Select("chat_id", "user_id", "token").Where(conds).ScanStruct(&u); !found || err != nil {
        return &u, false
    }

    return &u, true
}

func (me *Registrations) getChats(userIds []int) (res []int64) {
    me.db.From("registrations").Select("chat_id").Where(goqu.Ex{"user_id": userIds}).ScanVals(&res)
    return
}

func (me *Registrations) delete(userId int, chatId int64) {
    if _, err := me.db.From("registrations").Where(goqu.Ex{"user_id": userId, "chat_id": chatId}).Delete().Exec(); err != nil {
        log.Fatal(err)
    }
}

func (me *Registrations) add(tx *goqu.TxDatabase, reg *Registration) error {
    stmt := tx.From("registrations").Insert(goqu.Record{"user_id": reg.UserId, "chat_id": reg.ChatId, "token": reg.Token})
    if _, err := stmt.Exec(); err != nil {
        return err
    }

    return nil
}
