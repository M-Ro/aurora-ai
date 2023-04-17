package instance

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/M-Ro/aurora-ai/internal/discord"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "instance",
		Short: "serve the discord bot instance",
		Run:   Start,
	}
}

// Start the discord client
func Start(_ *cobra.Command, _ []string) {
	log.Info("Starting...")

	dg, err := discordgo.New("Bot " + viper.GetString("discord.auth_token"))
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	registerEventHandlers(dg)

	err = dg.Open()
	if err != nil {
		log.Error("Error opening Discord session: " + err.Error())
	}

	// Attach event handlers for exit
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func registerEventHandlers(dg *discordgo.Session) {
	dg.AddHandler(discord.OnReady)
	dg.AddHandler(discord.OnMessageCreate)
}
