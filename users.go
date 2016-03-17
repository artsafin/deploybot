package main

import (
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "sync"
    "fmt"
    "log"
    "gopkg.in/doug-martin/goqu.v3"
)

func NewRegistrations() *Registrations {
    return &Registrations{}
}

type TelegramUser struct {
    *tgbotapi.User
    data *User
}

type User struct {
    UserId int `db:"user_id"`
    ChatId int `db:"chat_id"`
    Token string `db:"token"`
    isReg bool
}

type Registrations struct {
    repo *Repository
    lock sync.RWMutex
}

func (me *Registrations) get(userId int, chatId int) (*User, bool) {
    me.lock.RLock()
    defer me.lock.RUnlock()

    u, err := me.repo.getRegistration(userId, chatId)

    if u == nil || err != nil {
        return &User{userId, chatId, "", false}, false
    }

    return u, true
}

func (me *Registrations) add(user *User, token string) error {
    me.lock.Lock()
    defer me.lock.Unlock()

    tokens := me.repo.getAvailableTokens()
    if !tokens.has(token) {
        return fmt.Errorf("Specified token is not available")
    }

    me.repo.addRegistration(user.UserId, user.ChatId, token)

    return nil
}

func (me *Repository) getChats(userIds []int) (res []int) {
    me.From("registrations").Select("chat_id").Where(goqu.Ex{"user_id": userIds}).ScanVals(&res)
    return
}

func (me *Repository) getRegistration(userId int, chatId int) (*User, error) {
    u := User{}

    conds := goqu.Ex{"user_id": userId, "chat_id": chatId}

    if found, err := me.From("registrations").Select("chat_id", "user_id", "token").Where(conds).ScanStruct(&u); !found || err != nil {
        return nil, err
    }

    u.isReg = true
    return &u, nil
}

func (me *Repository) deleteRegistration(userId int, chatId int) {
    if _, err := me.From("registrations").Where(goqu.Ex{"user_id": userId, "chat_id": chatId}).Delete().Exec(); err != nil {
        log.Fatal(err)
    }
}

func (me *Repository) addRegistration(userId int, chatId int, token string) {
    tx, err := me.Begin()
    if err != nil {
        log.Fatal(err)
    }

    err = tx.Wrap(func() error {
            stmt := tx.From("registrations").Insert(goqu.Record{"user_id": userId, "chat_id": chatId, "token": token})
            if _, err = stmt.Exec(); err != nil {
                return err
            }

            if err = me.markUsed(tx, token); err != nil {
                return err
            }

            return nil
        })

    
    if err != nil {
        log.Fatal(err)
    }
}
