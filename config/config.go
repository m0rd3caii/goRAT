package config

var (
	BotToken    string
	ChannelID   string
	PrivateChan string
)

func LoadConfig() {
	BotToken = ""
	ChannelID = ""
}
