package utils

var (
	LogLevel      = "info"
	DirectusHost  = "http://localhost:8055"
	DirectusToken = "directus-access-token"
	BotToken      = "my-bot-token"
)

const HELP_MESSAGE string = `This bot updates you on the singapore savings bonds interest rates! The following commands are available:
/subscribe adds you into the monthly ssb interest rate updates
/unsubscribe removes you from the monthly ssb interest rate updates
`
const DEFAULT_TIMEZONE = "Asia/Singapore"
