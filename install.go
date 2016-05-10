package main

import (
    "fmt"
    "github.com/artsafin/deploybot/data"
)

func install(dbSqlite string) {
    fmt.Println("Performing initial installation to " + dbSqlite)
    
    data.Init(dbSqlite)

    fmt.Println("Done installation to " + dbSqlite)
}
