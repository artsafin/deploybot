package main

import (
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "fmt"
    "strings"
)

type DeployBotAPI struct {
    *tgbotapi.BotAPI
}

func NewDeployBot(base *tgbotapi.BotAPI) DeployBotAPI {
    return DeployBotAPI{base}
}

type BotCtrl struct {
    state *State
    bot DeployBotAPI
}

func (ctrl *BotCtrl) ConsumeBotCommands(updates <-chan tgbotapi.Update) {
    for update := range updates {
        ctrl.processCommand(update)
    }
}

func (ctrl *BotCtrl) processCommand(update tgbotapi.Update) {
    fmt.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)

    msgParts := strings.Fields(update.Message.Text)

    if len(msgParts) == 0 {
        return
    }

    _, isReg := ctrl.state.regs.get(update.Message.From.ID)
    user := &User{&update.Message.From, update.Message.Chat.ID, isReg}

    switch msgParts[0] {
    case "/start":
        msg, err := ctrl.start(user, msgParts[1:])

        if err != nil {
            msg = err.Error()
        }

        ctrl.bot.Send(ctrl.createReplyMessage(update, msg, nil))
    case "/notify":
        msg, err := ctrl.notify(user, msgParts[1:])

        if err != nil {
            msg = err.Error()
        }

        ctrl.bot.Send(ctrl.createReplyMessage(update, msg, nil))
    case "/forget":
        msg, err := ctrl.forget(user, msgParts[1:])

        if err != nil {
            msg = err.Error()
        }
        ctrl.bot.Send(ctrl.createReplyMessage(update, msg, nil))
    case "/ping":
        msg := fmt.Sprintf("I am @%v, %v #%v and I send this message to any surviving bots", ctrl.bot.Self.UserName, ctrl.bot.Self.FirstName, ctrl.bot.Self.ID)
        ctrl.bot.Send(ctrl.createReplyMessage(update, msg, nil))
    default:
        ctrl.bot.Send(ctrl.createReplyMessage(update, "Sorry?", nil))
    }
}

func (ctrl *BotCtrl) createReplyMessage(update tgbotapi.Update, text string, configurator func(*tgbotapi.MessageConfig)) tgbotapi.MessageConfig {
    msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
    msg.ParseMode = "HTML"
    msg.ReplyToMessageID = update.Message.MessageID

    if configurator != nil {
        configurator(&msg)
    }

    return msg
}

func (ctrl *BotCtrl) start(user *User, args []string) (string, error) {
    if user.isReg {
        return "", fmt.Errorf("You've been already registered")
    }

    if len(args) != 1 {
        return "", fmt.Errorf("<code>token</code> is required. Obtain from developers.")
    }

    token := args[0]

    if !ctrl.state.cfg.hasToken(token) {
        return "", fmt.Errorf("Token not found")
    }

    err := ctrl.state.regs.add(user, token)
    if err != nil {
        return "", err
    }

    fmt.Printf("Registered user %v, token: %s\n", user, token)

    return fmt.Sprintf("%s, you have been registered", user.FirstName), nil
}

func (ctrl *BotCtrl) notify(user *User, args []string) (string, error) {
    if !user.isReg {
        return "", fmt.Errorf("You're not registered.\nFirst issue <code>/start token</code>")
    }

    if len(args) != 1 {
        return "", fmt.Errorf("<code>{service name}</code> is required\nCan be:\n" + ctrl.state.cfg.getServicesAsString())
    }

    service := args[0]

    if !ctrl.state.cfg.hasService(service) {
        return "", fmt.Errorf("Notification type is not supported")
    }

    ctrl.state.ns.subscribeUser(service, user)
    return fmt.Sprintf("You are now subscribed to %s events", service), nil
}

func (ctrl *BotCtrl) forget(user *User, args []string) (string, error) {
    if !user.isReg {
        return "", fmt.Errorf("You're not registered.\nFirst issue <code>/start token</code>")
    }

    if len(args) != 1 {
        return "", fmt.Errorf("<code>{service name}</code> is required\nCan be:\n" + ctrl.state.cfg.getServicesAsString())
    }

    service := args[0]

    ctrl.state.ns.unsubscribeUser(service, user)
    return fmt.Sprintf("You have been unsubscribed from %s events", service), nil
}