package context

import (
	"strings"

	"github.com/M-Ro/aurora-ai/internal/helpers"
	"github.com/spf13/viper"
)

type Author struct {
    Id      string
    Name    string
}

type ContextMessage struct {
    Author Author
    Message string
}

// NewCtxMsgFromBotResponse builds a new context message from the last bot response.
// The bot returns the entire conversation each time, so we need to isolate the last message.
func NewCtxMsgFromBotResponse(response string) ContextMessage {
    // We need to remove the last string from the human token
    // because the api is inconsistent and requires a suffix colon to inference without
    // going schizo, but at termination doesn't bother to produce a suffix colon itself.
    botToken := viper.GetString("llm.identifier_b")
    eosToken := viper.GetString("llm.identifier_p")
    eosToken = strings.TrimRight(eosToken, ":")

    lB := strings.LastIndex(response, botToken)
    lH := strings.LastIndex(response, eosToken)

    msg := ""
    // If lH > lB, the bot has re-prompted the user, so fetch the string upto that point
    if lH > lB {
        msg = helpers.Substr(
            response,
            lB + len(botToken),
            lH - (lB + len(botToken) + 1), // +1 to drop the \n the bot throws at the end
        )
    } else {
        msg = helpers.Substr(
            response,
            lB + len(botToken),
            len(response) - (lB + len(botToken)),
        )
    }

    return ContextMessage {
        Author: Author {
            Id: botToken,
            Name: botToken,
        },
        Message: msg,
    }
}
