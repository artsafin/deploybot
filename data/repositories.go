package data

import (
    "log"
    "fmt"
)

type Repositories struct {
    db *dataConn
    Registrations *Registrations
    AccessTokens *AccessTokens
    Notifications *Notifications
}

func NewRepositories(file string) *Repositories {
    db := newDataConn(file)

    return &Repositories{db, newRegistrations(db), newAccessTokens(db), newNotifications(db)}
}

func CloseRepositories(r *Repositories) {
    r.db.Db.Close()
}


func (me *Repositories) AddReg(reg *Registration) error {
    me.Registrations.lock.Lock()
    me.AccessTokens.lock.Lock()
    defer me.AccessTokens.lock.Unlock()
    defer me.Registrations.lock.Unlock()

    if me.AccessTokens.hasUnused(reg.Token) {
        tx, err := me.db.Begin()
        if err != nil {
            log.Fatal(err)
        }

        err = tx.Wrap(func() error {
            if err := me.Registrations.add(tx, reg); err != nil {
                return err
            }
            
            if err = me.AccessTokens.markUsed(tx, reg.Token); err != nil {
                return err
            }

            return nil
        })

        
        if err != nil {
            log.Fatal(err)
        }

        return nil
    } else {
        return fmt.Errorf("Unable to find specified token")
    }
}