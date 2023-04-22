package context

type ContextMap map[string]*ChatContext

// contexts is a singleton map containing all chat contexts
var contexts ContextMap = nil

// GetContext returns the ChatContext from the global state of all chat conversations.
// If the context cannot be found, a new one is created and returned.
func GetContext(key string) *ChatContext {
	if contexts == nil {
		contexts = make(ContextMap)
	}

	ctx, ok := contexts[key]
	if !ok {
		ctx = &ChatContext{}
		contexts[key] = ctx
	}

	return ctx
}
