package config

// Todo these aren't even app specific, move them into the api for each app

// Gradio function indices, which are no longer gauranteed to be constant
// This is fucking stupid, the packet indices vary between chat mode
const (
	// LLM indexes
	FnSetMode            = 19
	FnSetNamedMode       = 40
	FnSendPrepQuery      = 6
	FnSendNoFuckingIdea4 = 7
	FnSendNoFuckingIdea8 = 8 // In regular mode this apparently sets the parameters
	FnSendNoFuckingIdea9 = 9 // And this does the actual inference. Sometimes. Except on Tuesdays. On Tuesdays it returns random HTML

	// Stable diffusion indexes
	FnDoTheThing = 81
)
