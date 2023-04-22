package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	commands = []discordgo.ApplicationCommand{
		{
			Name:        "generate",
			Description: "Generate an image from a prompt via stable diffusion",
		},
	}
	commandsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"generate": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "generate_" + i.Interaction.Member.User.ID,
					Title:    "Stable Diffusion",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "positive_prompt",
									Label:       "Positive prompt. Specify desired traits",
									Style:       discordgo.TextInputParagraph,
									Placeholder: "puppy eaten by sharks, robot in background",
									Required:    true,
									MaxLength:   500,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "negative_prompt",
									Label:       "Negative prompt, specify undesired traits",
									Style:       discordgo.TextInputParagraph,
									Placeholder: "ugly, deformed, naked",
									Required:    false,
									MaxLength:   500,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "size",
									Label:       "Size. Only: 512x512, 512x768, 768x768.",
									Style:       discordgo.TextInputShort,
									Placeholder: "512x512",
									Required:    false,
									MaxLength:   500,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "seed",
									Label:       "Seed. Defaults -1 for random.",
									Style:       discordgo.TextInputShort,
									Placeholder: "-1",
									Required:    false,
									MaxLength:   500,
								},
							},
						},
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "Sampler",
									Label:       "Sampler: 'Euler A' or 'DPM++ 2M Karras'",
									Style:       discordgo.TextInputShort,
									Placeholder: "Euler A",
									Required:    false,
									MaxLength:   500,
								},
							},
						},
					},
				},
			})
			if err != nil {
				panic(err)
			}
		},
	}
)

func OnInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		if h, ok := commandsHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}

	case discordgo.InteractionModalSubmit:
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Generating request image...",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

		if err != nil {
			panic(err)
		}
		data := i.ModalSubmitData()

		if !strings.HasPrefix(data.CustomID, "generate") {
			return
		}

		msg, err := s.ChannelMessageSend(i.ChannelID, fmt.Sprintf(
			"Generation prompt from. From <@%s>\n**+ve**: %s\n**-ve**: %s",
			i.Member.User.ID,
			data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
			data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		))

		if err != nil {
			panic(err)
		}

		GenerateFromModalAndAttachMessage(s, msg, &data)
	}
}

func RegisterSlashCommands(s *discordgo.Session) {
	cmdIDs := make(map[string]string, len(commands))

	for _, cmd := range commands {
		rcmd, err := s.ApplicationCommandCreate(viper.GetString("discord.app_id"), "", &cmd)
		if err != nil {
			logrus.Fatalf("Cannot create slash command %q: %v", cmd.Name, err)
		}

		cmdIDs[rcmd.ID] = rcmd.Name
	}
}

func CleanupSlashCommands(s *discordgo.Session) {
	cmdIDs := make(map[string]string, len(commands))

	for id, name := range cmdIDs {
		err := s.ApplicationCommandDelete(viper.GetString("discord.app_id"), "", id)
		if err != nil {
			logrus.Fatalf("Cannot delete slash command %q: %v", name, err)
		}
	}
}
