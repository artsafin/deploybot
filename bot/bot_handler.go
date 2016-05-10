package bot

import (
	"strings"
    "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/artsafin/deploybot/data"
    "fmt"
)

func isFirstWord(text string, word string) bool {
    fields := strings.Fields(text)
    if len(fields) > 0 && (fields[0] == word || fields[0] == "/" + word) {
        return true
    }
    return false
}

type BotHandler struct {
    repo *data.Repositories
}

func NewBotHandler(repo *data.Repositories) *BotHandler {
    return &BotHandler{repo}
}

func (me *BotHandler) HandlerFunc(text string, uch *UserChat, update *tgbotapi.Update) (*HandlerReply, error) {
    if isFirstWord(text, "help") {
        return &HandlerReply{text: help(0)}, nil
    }

    if isFirstWord(text, "start") {
        return start(me.repo, text, uch)
    }

    if isFirstWord(text, "notify") {
        return notify(me.repo, text, uch.user)
    }

    if isFirstWord(text, "forget") {
        return forget(me.repo, text, uch.user)
    }

	return nil, BotErr("command not found")
}

func help(helpSection byte) string {
    head := `
Bot is able to notify you on currency rate changes.
All commands are available both with or without leading slash.

/help (or help) - Show this message
/instr (or instr) - Show supported instruments
/alerts (or alerts) - Show currently set alerts
/forget (or forget) - Unsubscribe from alert

`
    newa := `How to set new alerts:
    <code>/alert INSTRUMENT OPERATION VALUE</code>
    <code>alert INSTRUMENT OPERATION VALUE</code>
or just:
    <code>INSTRUMENT OPERATION VALUE</code>

where:
<code>INSTRUMENT</code> is one of the supported instruments (see /instr), case insensitive
<code>OPERATION</code> is one of:
    = (or eq, equals)
    &lt; (or lt, less than)
    &lt;= (or lte, less than or equals)
    &gt; (or gt, greater than)
    &gt;= (or gte, greater than or equals)
<code>VALUE</code> is a decimal number with dot as a decimal part separator

`
    if helpSection == 0 {
        return head + newa
    }

    var res string
    if helpSection & 1 == 1 {
        res += head
    }
    if helpSection & 2 == 2 {
        res += newa
    }
    return res
}

func start(repo *data.Repositories, text string, uch *UserChat) (*HandlerReply, error) {
    if repo.Registrations.Has(uch.user) {
        return nil, BotPublicErr("You've been already registered")
    }

    fields := strings.Fields(text)[1:]

    if len(fields) != 1 {
        return nil, BotPublicErr("<code>token</code> parameter is required. It should be obtained from developers.")
    }

    reg := &data.Registration{UserId: uch.user, ChatId: uch.chat, Token: fields[0]}

    err := repo.AddReg(reg)

    if err != nil {
        return nil, err
    }

    fmt.Printf("Registered user %v, token: %s\n", uch.user, fields[0])

    return &HandlerReply{text: fmt.Sprintf("You have been registered")}, nil
}

func notify(repo *data.Repositories, text string, userId int) (*HandlerReply, error) {
    if !repo.Registrations.Has(userId) {
        return nil, BotPublicErr("You're not registered.\nFirst issue <code>/start token</code>")
    }

    fields := strings.Fields(text)[1:]

    if len(fields) != 1 {
        return nil, BotPublicErr("<code>{service name}</code> is required")
    }

    err := repo.Notifications.Add(fields[0], userId)
    if err != nil {
        return nil, BotPublicErr(err.Error())
    }
    
    fmt.Printf("Notify user %v to %s\n", userId, fields[0])

    return &HandlerReply{text: fmt.Sprintf("You are now subscribed to %s events", fields[0])}, nil
}

func forget(repo *data.Repositories, text string, userId int) (*HandlerReply, error) {
    if !repo.Registrations.Has(userId) {
        return nil, BotPublicErr("You're not registered.\nFirst issue <code>/start token</code>")
    }

    fields := strings.Fields(text)[1:]

    if len(fields) != 1 {
        return nil, BotPublicErr("<code>{service name}</code> is required")
    }

    err := repo.Notifications.Delete(fields[0], userId)
    if err != nil {
        return nil, err
    }
    return &HandlerReply{text: fmt.Sprintf("You have been unsubscribed from %s events", fields[0])}, nil
}

