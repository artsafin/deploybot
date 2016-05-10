package main

import (
	"fmt"
    "os"
    "github.com/paked/configure"
    "github.com/artsafin/deploybot/bot"
    "github.com/artsafin/deploybot/event"
    "github.com/artsafin/deploybot/data"
    "strconv"
)

var (
    confPre = configure.New()
    configFilePath = confPre.String("config", "config.hcl", "Config file path")

    conf = configure.New()

    telegramToken  = conf.String("bot.token", "", "Telegram token")
    botDebug = conf.Bool("bot.debug", false, "Log network activity of bot")
    
    httpListen  = conf.String("http.listen", "localhost:8182", "Listen interface")
    dbSqlite  = conf.String("db.sqlite", "./deploybot.db", "Path to sqlite database for persistance")
    
    flagTestConfig = conf.Bool("testconfig", false, "Test config for errors and exit")
    flagInstall = conf.Bool("install", false, "Initialize database and exit")
    flagAddToken = conf.Bool("add-token", false, "Generate new token and exit. Usage: " + os.Args[0] + " <expires_in_sec>")
)

func init() {
    confPre.Use(configure.NewFlag())
    confPre.Parse()

    fmt.Println("Reading", *configFilePath)

    conf.Use(configure.NewEnvironment())
    conf.Use(configure.NewFlagWithUsage(usage))
    conf.Use(configure.NewHCLFromFile(*configFilePath))
}

func usage() string {
    return "USAGE"
}

func processSpecialModeAndExit() {
    if *flagTestConfig {
        fmt.Printf("OK\n%+v\n", conf)
        os.Exit(0)
    }

    if *flagInstall {
        install(*dbSqlite)
        os.Exit(0)
    }

    if *flagAddToken {
        if len(os.Args) < 3 || os.Args[2] == "" {
            fmt.Println("<expires_in_sec> argument required")
            os.Exit(1)
        }
        expiresIn, _ := strconv.Atoi(os.Args[2])
        repos := data.NewRepositories(*dbSqlite)
        token := repos.AccessTokens.AddRandom(int64(expiresIn))
        fmt.Println(token, "added")
        os.Exit(0)
    }
}

func main() {
	conf.Parse()
	fmt.Println("Deploy bot service, version:" + TAG)

    processSpecialModeAndExit()

    repo := data.NewRepositories(*dbSqlite)
    defer data.CloseRepositories(repo)

    botservice := bot.NewBotService(*telegramToken, *botDebug)
    fmt.Printf("Authorized on account %s\n", botservice.Bot().Self.UserName)

    botHandler := bot.NewBotHandler(repo)
    go botservice.ConsumeBotCommands(botHandler.HandlerFunc)

    listener := event.NewHttpListener()
    fmt.Println("Listening", *httpListen)
    listener.ListenEvents(*httpListen)
}

