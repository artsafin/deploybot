package main

import (
    "sync"
    "log"
    "gopkg.in/doug-martin/goqu.v3"
)

type Notifications struct {
    repo *Repository
    usersByService map[string]map[int]bool
    lock sync.RWMutex
}

func NewNotifications() Notifications {
    return Notifications{usersByService: make(map[string]map[int]bool)}
}

func LoadNotification(ns *Notifications, service string, userId int) {
    _, ok := ns.usersByService[service]
    if !ok {
        ns.usersByService[service] = make(map[int]bool)
    }

    ns.usersByService[service][userId] = true
}

func (me *Notifications) subscribeUser(service string, user *User) {
    me.lock.Lock()
    defer me.lock.Unlock()

    _, ok := me.usersByService[service]

    if !ok {
       me.usersByService[service] = make(map[int]bool)
    }

    me.usersByService[service][user.UserId] = true

    me.repo.addNotification(service, user.UserId)
}

func (me *Notifications) unsubscribeUser(service string, user *User) {
    me.lock.Lock()
    defer me.lock.Unlock()

    if _, ok := me.usersByService[service][user.UserId]; ok {
        delete(me.usersByService[service], user.UserId)

        me.repo.deleteNotification(service, user.UserId)
    }
}

func (me *Notifications) get(service string) (map[int]bool, bool) {
    me.lock.RLock()
    defer me.lock.RUnlock()

    um, ok := me.usersByService[service]

    if len(um) > 0 {
        return um, ok
    } else {
        return um, false
    }
}
func (me *Repository) getNotifications() *Notifications {
    ns := NewNotifications()

    var dbNs []struct{
        Service string `db:"service"`
        UserId  int `db:"user_id"`
    }

    if err := me.From("notifications").Select("service", "user_id").ScanStructs(&dbNs); err == nil {
        for _, dbN := range dbNs {
            LoadNotification(&ns, dbN.Service, dbN.UserId)
        }
    } else {
        log.Fatal(err)
    }

    ns.repo = me

    return &ns
}

func (me *Repository) deleteNotification(service string, userId int) {
    rec := goqu.Ex{"service": service, "user_id": userId}

    if _, err := me.From("notifications").Where(rec).Delete().Exec(); err != nil {
        log.Fatal(err)
    }
}

func (me *Repository) addNotification(service string, userId int) {
    rec := goqu.Record{"service": service, "user_id": userId}

    if _, err := me.From("notifications").Insert(rec).Exec(); err != nil {
        log.Fatal(err)
    }
}