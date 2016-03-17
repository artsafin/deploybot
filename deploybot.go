package main

import (
	"flag"
	"fmt"
	"log"
	"time"
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "os"
)

var (
    flagPingDuration = flag.Duration("botupdate", time.Second*5, "How often to ping Telegram for bot updates")
    flagBotDebug = flag.Bool("botdebug", false, "Log network activity of bot")
    flagTestConfig = flag.Bool("testconfig", false, "Test config for errors and exit")
    flagConfig = flag.String("config", "config.yml", "Path to config file")
    flagInstall = flag.Bool("install", false, "Initialize database and exit")
    flagAddToken = flag.Bool("add-token", false, "Generate new token and exit. Usage: " + os.Args[0] + " <expires_in_sec>")
)

type State struct {
    ns *Notifications
    regs *Registrations
    repo *Repository
    cfg *Config
}

func main() {
	flag.Parse()
	fmt.Println("Deploy bot service, version:" + TAG)

    cfgReader := NewConfigReader(*flagConfig)
    config, err := cfgReader.load()

    if err != nil {
        fmt.Printf("error: %v\n", err)
        os.Exit(1)
    }

    if *flagTestConfig {
        fmt.Printf("OK\n%+v\n", config)
        os.Exit(0)
    }

    if *flagInstall {
        install(config)
        os.Exit(0)
    }

    if *flagAddToken {
        if len(os.Args) < 3 || os.Args[2] == "" {
            fmt.Println("<expires_in_sec> argument required")
            os.Exit(1)
        }
        token := addToken(config, os.Args[2])
        fmt.Println(token, "added")
        os.Exit(0)
    }

    fmt.Println("Listening", config.Listen)

    repo := NewRepository(config.Db.Sqlite)
    defer repo.Db.Close()

    regs := NewRegistrations()
    regs.repo = repo
    state := &State{repo.getNotifications(), regs, repo, config}

	baseBot, err := tgbotapi.NewBotAPI(config.Telegram_token)
	if err != nil {
		log.Panic(err)
	}

	bot := NewDeployBot(baseBot)

	bot.Debug = *flagBotDebug

	fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = int(flagPingDuration.Seconds())

	updatesChan, err := bot.GetUpdatesChan(u)

    if err != nil {
        log.Panic(err)
    }

    botCtrl := BotCtrl{state, bot}

    go botCtrl.ConsumeBotCommands(updatesChan)

    eventChan := NewEventChan()

    go ListenEvents(config.Listen, eventChan)

    ConsumeEvents(state, bot, eventChan)
}

