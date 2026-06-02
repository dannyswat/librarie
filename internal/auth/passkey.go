package auth

import (
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5/pgtype"

	"librarie/internal/db"
)

// WebAuthnUser wraps a db.User to satisfy the webauthn.User interface.
type WebAuthnUser struct {
	User        db.User
	Credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	id := u.User.ID
	return id.Bytes[:]
}

func (u *WebAuthnUser) WebAuthnName() string {
	return u.User.Username
}

func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.User.Username
}

func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

// WebAuthnIcon is deprecated and unused.
func (u *WebAuthnUser) WebAuthnIcon() string { return "" }

// dbPasskeyToCredential converts a db.Passkey to a webauthn.Credential.
func dbPasskeyToCredential(p db.Passkey) webauthn.Credential {
	return webauthn.Credential{
		ID:        p.CredentialID,
		PublicKey: p.PublicKey,
		Authenticator: webauthn.Authenticator{
			SignCount: uint32(p.SignCount),
		},
	}
}

// passkeyFromCredential converts a webauthn.Credential back to DB-storable fields.
type newPasskeyFields struct {
	CredentialID []byte
	PublicKey    []byte
	SignCount    int64
	UserID       pgtype.UUID
}

func credentialToPasskeyFields(userID pgtype.UUID, cred *webauthn.Credential) newPasskeyFields {
	return newPasskeyFields{
		CredentialID: cred.ID,
		PublicKey:    cred.PublicKey,
		SignCount:    int64(cred.Authenticator.SignCount),
		UserID:       userID,
	}
}

// challengeStore holds in-memory WebAuthn session data keyed by a challenge cookie value.
type challengeStore struct {
	mu      sync.Mutex
	entries map[string]challengeEntry
}

type challengeEntry struct {
	data    *webauthn.SessionData
	expires time.Time
}

var globalChallengeStore = &challengeStore{
	entries: make(map[string]challengeEntry),
}

func (s *challengeStore) Set(key string, data *webauthn.SessionData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[key] = challengeEntry{data: data, expires: time.Now().Add(5 * time.Minute)}
	// Evict stale entries inline (cheap enough for low traffic).
	for k, v := range s.entries {
		if time.Now().After(v.expires) {
			delete(s.entries, k)
		}
	}
}

func (s *challengeStore) Get(key string) (*webauthn.SessionData, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[key]
	if !ok || time.Now().After(e.expires) {
		delete(s.entries, key)
		return nil, false
	}
	delete(s.entries, key) // single-use
	return e.data, true
}

// challengeCookieName is the cookie used to link begin→complete calls.
const challengeCookieName = "librarie_wac"
