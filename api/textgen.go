package api

type QueryRequest struct {
    Context     string          `json:"context"`
    Parameters  ParameterSet
}

type SendHashRequest struct {
    SessionHash string  `json:"session_hash"`
    FnIndex     uint32  `json:"fn_index"`
}

type SendDataRequest struct {
    SessionHash string      `json:"session_hash"`
    FnIndex     uint32      `json:"fn_index"`
    Data        []string    `json:"data"`
    EventData   *string     `json:"event_data"`
}

type ParameterSet struct {
    MaxNewTokens             uint32 `json:"max_new_tokens"`
    MaximumPromptTokens      uint32 `json:"maximum_prompt_tokens"`
    Temperature              uint32 `json:"temperature"`
    TopP                     uint32 `json:"top_p"`
    TopK                     uint32 `json:"top_k"`
    TypicalP                 uint32 `json:"typical_p"`
    RepetitionPenalty        uint32 `json:"repetition_penalty"`
    EncoderRepetitionPenalty uint32 `json:"encoder_repetition_penalty"`
    NoRepeatNgramSize        uint32 `json:"no_repeat_ngram_size"`
    MinLength                uint32 `json:"min_length"`
    DoSample                 bool   `json:"do_sample"`
    Seed                     int64  `json:"seed"`
}

type GradioResponsePacket struct {
    Message      GradioResponseMessage  `json:"msg"`
    Output       *GradioResponseOutput  `json:"output"`
    Success      *bool                  `json:"success"`
}

type GradioResponseOutput struct {
    Data            []string    `json:"data"`
    IsGenerating    bool        `json:"is_generating"`
    Duration        float32     `json:"duration"`
    AverageDuration float32     `json:"average_duration"`
}

// Gradio response message types, this is essentially a packet type enum
type GradioResponseMessage string
const (
    MsgSendHash             GradioResponseMessage = "send_hash"
    MsgSendData             GradioResponseMessage = "send_data"
    MsgEstimation           GradioResponseMessage = "estimation"
    MsgProcessStarts        GradioResponseMessage = "process_starts"
    MsgProcessGenerating    GradioResponseMessage = "process_generating"
    MsgProcessCompleted     GradioResponseMessage = "process_completed"
)
