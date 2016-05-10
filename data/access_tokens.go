package data

import (
    "sync"
    "log"
    "fmt"
    "gopkg.in/doug-martin/goqu.v3"
    "time"
    "crypto/rand"
    "crypto/sha1"
)

func newAccessTokens(db *dataConn) *AccessTokens {
    return &AccessTokens{db: db}
}

type AccessTokens struct {
    db *dataConn
    lock sync.RWMutex
}

func (me *AccessTokens) AddRandom(expiresSec int64) string {
    me.lock.Lock()
    defer me.lock.Unlock()

    rndBytes := make([]byte, 100)

    if _, err := rand.Read(rndBytes); err != nil {
        log.Fatal(err)
    }
    newToken := fmt.Sprintf("%x", sha1.Sum(rndBytes))

    expiresUnix := time.Now().Unix() + expiresSec

    rec := goqu.Record{"token": newToken, "expires": expiresUnix, "used": 0,}

    _, err := me.db.From("access_tokens").Insert(rec).Exec()

    if err != nil {
        log.Fatal(err)
    }

    return newToken
}

func (me *AccessTokens) hasUnused(k string) bool {
    conds := goqu.Ex{"used": 0, "expires": goqu.Op{"gt": time.Now().Unix()}, "token": k}
    count, err := me.db.From("access_tokens").Select("token").Where(conds).Count()
    if err != nil {
        log.Fatal(err)
    }

    return count > 0
}

func (me *AccessTokens) markUsed(tx *goqu.TxDatabase, token string) error {
    stmt := tx.From("access_tokens").Where(goqu.Ex{"used": 0, "token": token}).Update(goqu.Record{"used": 1})
    result, err := stmt.Exec()

    if err != nil {
        return err
    }

    affected, err := result.RowsAffected()
    if err != nil {
        return err
    }

    if affected != 1 {
        return fmt.Errorf("AccessTokens::markUsed failed")
    }

    return nil
}

/*
func (me *Repository) getAvailableTokens() Tokens {
    var tokens []string

    var conds = goqu.Ex{"used": 0, "expires": goqu.Op{"gt": time.Now().Unix()}}
    if err := me.From("access_tokens").Select("token").Where(conds).ScanVals(&tokens); err != nil {
        log.Fatal(err)
    }

    m := make(map[string]bool, len(tokens))
    for _, v := range tokens {
        m[v] = true
    }

    return Tokens(m)
}
*/

