package main

import "sync"

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

    me.usersByService[service][user.ID] = true

    me.repo.addNotification(service, user.ID)
}

func (me *Notifications) unsubscribeUser(service string, user *User) {
    me.lock.Lock()
    defer me.lock.Unlock()

    if _, ok := me.usersByService[service][user.ID]; ok {
        delete(me.usersByService[service], user.ID)

        me.repo.deleteNotification(service, user.ID)
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