package main

import (
    "github.com/go-telegram-bot-api/telegram-bot-api"
    "fmt"
    "time"
    "net/http"
    "io"
    "log"
)

type event struct {
    service string
    tag string
    event string
    comment string
    ts time.Time
}

func NewEventChan() chan event {
    return make(chan event, 10)
}

func parseTime(str string) time.Time {
    timeVariants := []string{"2006-01-02 15:04:05 MST", "2006-01-02 15:04:05",}

    var evTs time.Time
    var timeErr error

    for i:=0; i<len(timeVariants) && (i == 0 || timeErr != nil); i++ {
        evTs, timeErr = time.Parse(timeVariants[i], str)
    }

    if timeErr != nil {
        evTs = time.Now()
    }

    return evTs
}

func newServerHandler(ch chan<- event) func (w http.ResponseWriter, req *http.Request) {
    return func (w http.ResponseWriter, req *http.Request) {
        req.ParseForm()

        ev := event{req.Form.Get("service"), req.Form.Get("tag"), req.Form.Get("event"), req.Form.Get("comment"), parseTime(req.Form.Get("ts"))}

        if ev.service == "" {
            log.Printf("Request %s lacks 'service' parameter\n", req.RequestURI)
        }
        ch <- ev

        io.WriteString(w, "OK\n")
    }
}

func ListenEvents(bindTo string, ch chan<- event) {
    http.HandleFunc("/deploy", newServerHandler(ch))
    http.ListenAndServe(bindTo, nil)
}

func ConsumeEvents(state *State, bot DeployBotAPI, ch <-chan event) {
    for v := range ch {
        log.Println("event:", v)

        if userIDs, ok := state.ns.get(v.service); ok {
            onEvent(state.regs, userIDs, bot, v)
        }
    }
}

func onEvent(regs *Registrations, userIdsMap map[int]bool, bot DeployBotAPI, alert event) {
    log.Println("Registered alert:", alert)

    var text string
    if len(alert.comment) > 0 {
        text = fmt.Sprintf("<b>%s</b> (<b>%s</b>) ref %s has been %s at %v", alert.service, alert.comment, alert.tag, alert.event, alert.ts)
    } else {
        text = fmt.Sprintf("<b>%s</b> ref %s has been %s at %v", alert.service, alert.tag, alert.event, alert.ts)
    }

    userIds := make([]int, len(userIdsMap))
    for userId := range userIdsMap {
        userIds = append(userIds, userId)
    }

    chatIds := regs.repo.getChats(userIds)

    for _, chatId := range chatIds {
        msg := tgbotapi.NewMessage(chatId, text)
        msg.ParseMode = "HTML"
        bot.Send(msg)
    }
}
