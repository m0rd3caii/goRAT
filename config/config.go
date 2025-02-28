package config

var (
	BotToken  string
	ChannelID string
)

func LoadConfig() {
	BotToken = ""
	ChannelID = ""
}
