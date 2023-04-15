package discord

import (
	"github.com/M-Ro/aurora-ai/internal/textgen/context"
	"github.com/bwmarrin/discordgo"
)

// NewCtxMsgFromDiscordMsg builds a new context message from a discord message.
func NewCtxMsgFromDiscordMsg(msg *discordgo.Message) context.ContextMessage {
    return context.ContextMessage {
        Author: context.Author {
            Id: msg.Author.ID,
            Name: msg.Author.Username,
        },
        Message: msg.ContentWithMentionsReplaced(),
    }
}
