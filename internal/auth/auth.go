package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	codeFile       = ".accesscode"
	sessionTimeout = 30 * time.Minute
)

// Manager handles access code storage and verification.
type Manager struct {
	mu       sync.RWMutex
	sessions map[string]time.Time // token → last activity time
}

func NewManager() *Manager {
	return &Manager{sessions: make(map[string]time.Time)}
}

// IsSetup returns true if an access code has been set.
func (m *Manager) IsSetup() bool {
	_, err := os.Stat(codeFile)
	return err == nil
}

// Setup hashes and stores the access code. Fails if already set.
func (m *Manager) Setup(code string) error {
	if m.IsSetup() {
		return fmt.Errorf("access code already set")
	}
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("access code cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash access code: %w", err)
	}

	if err := os.WriteFile(codeFile, hash, 0o600); err != nil {
		return fmt.Errorf("write access code: %w", err)
	}
	return nil
}

// Verify checks the access code and returns a session token if valid.
func (m *Manager) Verify(code string) (string, error) {
	hash, err := os.ReadFile(codeFile)
	if err != nil {
		return "", fmt.Errorf("read access code: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte(code)); err != nil {
		return "", fmt.Errorf("invalid access code")
	}

	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}

	m.mu.Lock()
	m.sessions[token] = time.Now()
	m.mu.Unlock()

	return token, nil
}

// ValidSession returns true if the token is a valid, non-expired session.
func (m *Manager) ValidSession(token string) bool {
	m.mu.RLock()
	lastActivity, ok := m.sessions[token]
	m.mu.RUnlock()

	if !ok {
		return false
	}
	if time.Since(lastActivity) > sessionTimeout {
		// Expired — clean up
		m.mu.Lock()
		delete(m.sessions, token)
		m.mu.Unlock()
		return false
	}
	return true
}

// TouchSession extends the session's last activity time.
func (m *Manager) TouchSession(token string) {
	m.mu.Lock()
	if _, ok := m.sessions[token]; ok {
		m.sessions[token] = time.Now()
	}
	m.mu.Unlock()
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
