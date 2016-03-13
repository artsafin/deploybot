package main

import (
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "sync"
    "fmt"
)

type User struct {
    *tgbotapi.User
    chatId int
    isReg bool
}

func NewRegistrations() Registrations {
    return Registrations{users: make(map[int]*User), takenTokens: make(map[string]bool)}
}

type Registrations struct {
    users map[int]*User
    takenTokens map[string]bool
    lock sync.RWMutex
}

func (me *Registrations) get(userId int) (*User, bool) {
    me.lock.RLock()
    defer me.lock.RUnlock()

    u, regOk := me.users[userId]

    return u, regOk
}

func (me *Registrations) add(user *User, token string) error {
    me.lock.Lock()
    defer me.lock.Unlock()

    if _, tokenTaken := me.takenTokens[token]; tokenTaken {
        return fmt.Errorf("Token already taken")
    }

    me.users[user.ID] = user
    me.takenTokens[token] = true

    return nil
}