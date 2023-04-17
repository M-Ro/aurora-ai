package discord

import (
	"strings"
	"time"

	"github.com/M-Ro/aurora-ai/internal/textgen"
	"github.com/M-Ro/aurora-ai/internal/textgen/context"
	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
)

func OnReady(s *discordgo.Session, event *discordgo.Ready) {
	logrus.Info("Discord ready")
}

func OnMessageCreate(s *discordgo.Session, msg *discordgo.MessageCreate) {
	// Ignore our own messages
	if msg.Author.ID == s.State.User.ID {
		return
	}

    // Check if this was a command.
    if handleTextCommand(s, msg) {
        return
    }

	// Ignore if we weren't @mentioned
	found := false
	for _, user := range msg.Mentions {
		if user.ID == s.State.User.ID {
			found = true
			break
		}
	}

	if !found {
		return
	}

	// Flag that we are generating a response
	err := s.ChannelTyping(msg.ChannelID)
	if err != nil {
		logrus.Error("Failed to set typing state: " + err.Error())
	}

	respond(s, msg)
}

func respond(s *discordgo.Session, msg *discordgo.MessageCreate) {
	lastTime := time.Now().UnixMilli()

    // Get the chat ctx for this channel, build & append a new ctx msg from the discord msg
    chatCtx := context.GetContext(msg.ChannelID)
    ctxMsg := NewCtxMsgFromDiscordMsg(s, msg)
    err := chatCtx.AddMessage(&ctxMsg)
    if err != nil {
        logrus.Error("Failed to add message to chat context.", err)
        return
    }

    // Run inference, update discord message as we get new tokens.
	var sendMsg *discordgo.Message
	err = textgen.RunInference(
        chatCtx.Prompt(), // Send the entire compiled conversation prompt to the inferencer
		func(output string) {
            if len(output) <= 0 {
                return
            }

            ctxBotResponseMsg := context.NewCtxMsgFromBotResponse(output)
            if len(ctxBotResponseMsg.Message) <= 0 {
                return
            }

			if sendMsg == nil {
				sendMsg, err = s.ChannelMessageSend(msg.ChannelID, ctxBotResponseMsg.Message)
				if err != nil {
					logrus.Error("fek", err)
					return
				}
			} else {
				if time.Now().UnixMilli() > lastTime+750 {
					lastTime = time.Now().UnixMilli()
					sendMsg, err = s.ChannelMessageEdit(msg.ChannelID, sendMsg.Reference().MessageID, ctxBotResponseMsg.Message)
					if err != nil {
						logrus.Error("fek", err)
						return
					}
				}
			}
		},
		func(output string) {
            if len(output) <= 0 {
                return
            }

            ctxBotResponseMsg := context.NewCtxMsgFromBotResponse(output) 

			if sendMsg != nil {
				sendMsg, err = s.ChannelMessageEdit(msg.ChannelID, sendMsg.Reference().MessageID, ctxBotResponseMsg.Message)
				if err != nil {
					logrus.Error("fek", err)
					return
				}

                // Add the message to the convo prompt
                chatCtx.AddMessage(&ctxBotResponseMsg)
			}
		},
	)
}

type TextCommandFunction func(*discordgo.Session, *discordgo.MessageCreate)
type CommandFuncMap map[string]TextCommandFunction

var commands CommandFuncMap = nil 

// Checks if the message matches any known text command. If so, 
func handleTextCommand(s *discordgo.Session, msg *discordgo.MessageCreate) bool {
    if commands == nil {
        commands = make(CommandFuncMap)
        registerTextCommands(commands)
    }

    for key, callFunc := range commands {
        if strings.HasPrefix(msg.Content, key) {
            callFunc(s, msg)
            return true
        }
    }

    return false;
}

func registerTextCommands(commands CommandFuncMap) {
    commands["!reset"] = cmdResetConversation
}

func cmdResetConversation(s *discordgo.Session, msg *discordgo.MessageCreate) {
    chatCtx := context.GetContext(msg.ChannelID)

    if chatCtx == nil {
        return
    }

    chatCtx.Messages = []context.ContextMessage{}

    _, err := s.ChannelMessageSend(msg.ChannelID, "Conversation has been reset.")
    if err != nil {
        logrus.Error("fek", err)
    }
}
