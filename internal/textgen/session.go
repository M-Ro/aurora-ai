package textgen

import "github.com/google/uuid"

type Session struct {
    SessionHash   string  `json:"session_hash"`
}

var globalSession *Session = nil

func newSession() *Session {
    s := Session {
        SessionHash: quickHash(),
    }

    return &s
}

// TODO: Serialize the hash for future use if the bot restarts, so we can maintain
// context.
func quickHash() string {
    return (uuid.New()).String();
}

func GetSession() *Session {
    if globalSession == nil {
        globalSession = newSession()
    }

    return globalSession
}
