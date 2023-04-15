package context

import (
	"strings"

	"github.com/spf13/viper"
)

type ChatContext struct {
    Messages    []ContextMessage
}

// EnforceSize truncates old messages so we don't go over the token limit.
func (c *ChatContext) EnforceSize() {
    limit := viper.GetInt("llm.settings.maximum_prompt_tokens")

    for c.TokenCount() > limit {
        c.Messages = c.Messages[1:]
    }
}

// Adds a chat message to the prompt.
func (c *ChatContext) AddMessage(ctxMsg *ContextMessage) error {
    c.Messages = append(c.Messages, *ctxMsg)
    c.EnforceSize()

    return nil
}

// Prompt returns the current conversation prompt.
func (c *ChatContext) Prompt() string {
    return ""
}

// CalculateTokenCount gets the token count from the context.
// HACK: this just counts words since we don't have access
// to the tokenizer output.
func (c *ChatContext) TokenCount() int {
    prompt := c.Prompt()

    return len(strings.Fields(prompt))
}
