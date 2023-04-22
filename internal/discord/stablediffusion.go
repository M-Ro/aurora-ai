package discord

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/M-Ro/aurora-ai/internal/stablediffusion"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type discordModalParams struct {
	PositivePrompt string
	NegativePrompt string
	Size           string
	Seed           string
	Sampler        string
}

func newDiscordModalParams(data *discordgo.ModalSubmitInteractionData) *discordModalParams {
	params := discordModalParams{
		PositivePrompt: data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		NegativePrompt: data.Components[1].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		Size:           data.Components[2].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		Seed:           data.Components[3].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
		Sampler:        data.Components[4].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value,
	}

	return &params
}

func ParameterSetFromDiscordParams(params *discordModalParams) stablediffusion.ParameterSet {
	set := stablediffusion.NewParameterSet()

	// Inject variables from discord command
	set.PositivePrompt = params.PositivePrompt
	set.NegativePrompt = params.NegativePrompt

	dimensions, err := getSizeFromDiscordModalVal(params.Size)
	if err == nil && isValidDimensions(dimensions[0], dimensions[1]) {
		set.Width = dimensions[0]
		set.Height = dimensions[1]
	}

	paramSeed, err := strconv.Atoi(params.Seed)
	if err == nil && paramSeed > 0 {
		set.Seed = int32(paramSeed)
	}

	sampler, err := asValidSampler(params.Sampler)
	if err == nil {
		set.SampleMethod = sampler
	}

	return set
}

var (
	ErrInvalidValue = errors.New("Invalid value from modal form")
)

// getSizeFromDiscordModalVal returns a pair of uint32's from a string containing image dimensions.
// e.g "512 x 768" => [512, 768]
// on parse failure, an empty slice will be returned with ErrInvalidValue
func getSizeFromDiscordModalVal(val string) ([]uint32, error) {
	size := strings.TrimSpace(val)

	dimensions := strings.Split(size, "x")

	if len(dimensions) != 2 {
		return []uint32{}, ErrInvalidValue
	}

	x, err := strconv.Atoi(dimensions[0])
	if err != nil {
		return []uint32{}, ErrInvalidValue
	}

	y, err := strconv.Atoi(dimensions[1])
	if err != nil {
		return []uint32{}, ErrInvalidValue
	}

	return []uint32{
		uint32(x), uint32(y),
	}, nil
}

// isValidDimensions returns whether the dimensions provided are valid within
// the contraints of the stable diffusion model.
func isValidDimensions(x uint32, y uint32) bool {
	min := uint32(512)
	max := uint32(768)

	if x < min || y < min || x > max || y > max {
		return false
	}

	return true
}

func asValidSampler(sampler string) (string, error) {
	// TODO add the rest
	validSamplers := []string{
		"Euler a",
		"Euler",
		"DPM++ 2M Karras",
	}

	for _, validSampler := range validSamplers {
		if strings.ToLower(validSampler) == strings.ToLower(sampler) {
			return validSampler, nil
		}
	}

	return "", errors.New("Invalid sampler")
}

func GenerateFromModalAndAttachMessage(
	s *discordgo.Session,
	msg *discordgo.Message,
	data *discordgo.ModalSubmitInteractionData,
) {
	discordParams := newDiscordModalParams(data)
	sdParams := ParameterSetFromDiscordParams(discordParams)

	err := stablediffusion.Run(
		&sdParams,
		func(images []bytes.Reader, err error) {
			if err != nil {
				setErrorMessage(s, msg, err)
				return
			}

			embeds, files, err := getDiscordAttachmentsFromSdImages(images)
			if err != nil {
				setErrorMessage(s, msg, err)
				return
			}

			messageEdit := discordgo.NewMessageEdit(msg.ChannelID, msg.ID)
			messageEdit.Content = &msg.Content
			messageEdit.Embeds = embeds
			messageEdit.Files = files
			_, err = s.ChannelMessageEditComplex(messageEdit)

			if err != nil {
				logrus.Error("Failed editing message with attachments")
			}
		},
	)

	if err != nil {
		setErrorMessage(s, msg, err)
	}
}

var (
	ErrNoImagesToAttach = errors.New("No images to attach to channel message")
)

func getDiscordAttachmentsFromSdImages(
	images []bytes.Reader,
) ([]*discordgo.MessageEmbed, []*discordgo.File, error) {
	embeds := []*discordgo.MessageEmbed{}
	files := []*discordgo.File{}

	if len(images) == 0 {
		return embeds, files, ErrNoImagesToAttach
	}

	for _, reader := range images {
		filename := fmt.Sprintf("%s.png", (uuid.New()).String())

		embeds = append(embeds, &discordgo.MessageEmbed{
			Image: &discordgo.MessageEmbedImage{
				URL: fmt.Sprintf("attachment://" + filename),
			},
		})

		files = append(files, &discordgo.File{
			Name:   filename,
			Reader: &reader,
		})
	}

	return embeds, files, nil
}

func setErrorMessage(s *discordgo.Session, msg *discordgo.Message, err error) {
	content := fmt.Sprintf("%s\n%s", msg.Content, err)
	_, editErr := s.ChannelMessageEdit(msg.ChannelID, msg.ID, content)
	if editErr != nil {
		logrus.Error("Ironic error outputting error message for error: %s", err)
	}
}
