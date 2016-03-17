package main

import (
    "log"
    "gopkg.in/doug-martin/goqu.v3"
    "time"
)

type Tokens map[string]bool

func (me *Tokens) has(k string) bool {
    _, ok := (*me)[k]

    return ok
}

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

func (me *Repository) markUsed(tx *goqu.TxDatabase, token string) error {
    stmt := tx.From("access_tokens").Where(goqu.Ex{"token": token}).Update(goqu.Record{"used": 1})
    _, err := stmt.Exec()

    return err
}
