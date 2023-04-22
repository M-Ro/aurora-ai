package discord

import (
	"fmt"
	"strings"

	"github.com/M-Ro/aurora-ai/internal/textgen/context"
	"github.com/bwmarrin/discordgo"
)

// NewCtxMsgFromDiscordMsg builds a new context message from a discord message.
func NewCtxMsgFromDiscordMsg(s *discordgo.Session, msg *discordgo.MessageCreate) context.ContextMessage {
	// Remove the @ to the bot.
	token := fmt.Sprintf("@%s ", s.State.User.Username)
	queryContent := strings.Replace(msg.ContentWithMentionsReplaced(), token, "", -1)

	return context.ContextMessage{
		Author: context.Author{
			Id:   msg.Author.ID,
			Name: msg.Author.Username,
		},
		Message: queryContent,
	}
}
