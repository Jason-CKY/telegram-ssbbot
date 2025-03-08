package utils

var (
	LogLevel             string
	DirectusHost         string
	DirectusToken        string
	BotToken             string
	WhitelistedUsernames []string
)

const HELP_MESSAGE string = `This bot updates you on the singapore savings bonds interest rates! The following commands are available:
/subscribe adds you into the monthly ssb interest rate updates
/unsubscribe removes you from the monthly ssb interest rate updates
`
const DEFAULT_TIMEZONE = "Asia/Singapore"
