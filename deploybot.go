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
	flagTokens = flag.String("tokens", "access_tokens.yml", "Path to tokens file")
)

type State struct {
    ns *Notifications
    regs *Registrations
    cfg *Config
}

func main() {
	flag.Parse()
	fmt.Println("Deploy bot service v." + TAG)

    cfgReader := NewConfigReader(*flagConfig, *flagTokens)
    config, err := cfgReader.load()

    if err != nil {
        fmt.Printf("error: %v\n", err)
        os.Exit(1)
    }

    if *flagTestConfig {
        fmt.Printf("OK\n%+v\n", config)
        os.Exit(0)
    }

    ns, regs := NewNotifications(), NewRegistrations()
    state := &State{&ns, &regs, config}

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

