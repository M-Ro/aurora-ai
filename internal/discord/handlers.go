package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/M-Ro/aurora-ai/internal/textgen"
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
	var sendMsg *discordgo.Message
	var err error

	content := ""
	lastTime := time.Now().UnixMilli()

    // hack: replace the bot @ with the identifier token
    token := fmt.Sprintf("<@%s> ", s.State.User.ID)
    queryContent := strings.Replace(msg.Message.Content, token, "", -1)

	err = textgen.RunInference(
        queryContent,
		func(output string) {
			content = output

            if len(content) <= 0 {
                return
            }

			if sendMsg == nil {
				sendMsg, err = s.ChannelMessageSend(msg.ChannelID, content)
				if err != nil {
					logrus.Error("fek", err)
					return
				}
			} else {
				if time.Now().UnixMilli() > lastTime+500 {
					lastTime = time.Now().UnixMilli()
					sendMsg, err = s.ChannelMessageEdit(msg.ChannelID, sendMsg.Reference().MessageID, output)
					if err != nil {
						logrus.Error("fek", err)
						return
					}
				}
			}
		},
		func() {
			if sendMsg != nil {
				sendMsg, err = s.ChannelMessageEdit(msg.ChannelID, sendMsg.Reference().MessageID, content)
				if err != nil {
					logrus.Error("fek", err)
					return
				}
			}
		},
	)

}
