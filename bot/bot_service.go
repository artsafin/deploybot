package bot

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strings"
)

type BotSvc struct {
	bot *tgbotapi.BotAPI
}

type HandlerReply struct {
	text         string
	configurator func(*tgbotapi.MessageConfig)
}

type HandlerCb func(string, *UserChat, *tgbotapi.Update) (*HandlerReply, error)

type UserChat struct {
    user int
    chat int64
    // user *tgbotapi.User
    // chat *tgbotapi.Chat
}

func NewBotService(token string, isDebug bool) *BotSvc {
	baseBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	baseBot.Debug = isDebug

	return &BotSvc{baseBot}
}

func (svc *BotSvc) Bot() *tgbotapi.BotAPI {
    return svc.bot
}

func (svc *BotSvc) ConsumeBotCommands(handler HandlerCb) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 2

	updatesChan, err := svc.bot.GetUpdatesChan(u)

	if err != nil {
		log.Panic(err)
	}

	for update := range updatesChan {
		svc.processCommand(&update, handler)
	}
}

func (svc *BotSvc) processCommand(update *tgbotapi.Update, handler HandlerCb) {
	fmt.Printf("[%s] %s\n", update.Message.From.UserName, update.Message.Text)

	uch := UserChat{update.Message.From.ID, update.Message.Chat.ID}

	reply, err := handler(update.Message.Text, &uch, update)

	if err == nil {
        if reply != nil {
		  svc.bot.Send(svc.createReplyMessage(update, reply.text, reply.configurator))
        }
		return
	}
    if _, ok := err.(BotPublicErr); ok {
        svc.bot.Send(svc.createReplyMessage(update, err.Error(), nil))
        return
    }

	svc.replyDefault(update)
}

func (svc *BotSvc) replyDefault(update *tgbotapi.Update) {
    msgParts := strings.Fields(update.Message.Text)

    if len(msgParts) == 0 {
        return
    }

    switch msgParts[0] {
    case "/ping":
        msg := fmt.Sprintf("I am @%v, %v #%v", svc.bot.Self.UserName, svc.bot.Self.FirstName, svc.bot.Self.ID)
        svc.bot.Send(svc.createReplyMessage(update, msg, nil))
    default:
        svc.bot.Send(svc.createReplyMessage(update, "Sorry?", nil))
    }
}

func (svc *BotSvc) createReplyMessage(update *tgbotapi.Update, text string, configurator func(*tgbotapi.MessageConfig)) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	// msg.ReplyToMessageID = update.Message.MessageID

	if configurator != nil {
		configurator(&msg)
	}

	return msg
}
