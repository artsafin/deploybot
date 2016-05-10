package bot

type BotErr string

func (e BotErr) Error() string {
    return string(e)
}

type BotPublicErr string

func (e BotPublicErr) Error() string {
    return string(e)
}
