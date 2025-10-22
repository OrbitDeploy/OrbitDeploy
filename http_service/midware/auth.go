package midware

import (
	"time"
)

// SessionData represents user session information
type SessionData struct {
	UserID    uint
	Username  string
	ExpiresAt time.Time
}

// sessions stores active user sessions (in production, use a proper session store)
var sessions = make(map[string]*SessionData)

// CheckAuthCompat provides compatibility with http.HandlerFunc for handlers not yet converted

// GetSessions returns the sessions map for use by auth handlers
func GetSessions() map[string]*SessionData {
	return sessions
}

// SetSession stores a new session
func SetSession(token string, session *SessionData) {
	sessions[token] = session
}

// DeleteSession removes a session
func DeleteSession(token string) {
	delete(sessions, token)
}
