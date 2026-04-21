package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"notebook-mcp/internal/repo"

	"golang.org/x/crypto/bcrypt"
)

type Config struct {
	IssuerURL      string
	ClientID       string
	CodeTTL        time.Duration
	AccessTokenTTL time.Duration
}

type Service struct {
	cfg    Config
	users  *repo.UserRepo
	mu     sync.RWMutex
	codes  map[string]authCode
	tokens map[string]accessToken
}

type authCode struct {
	UserID              int64
	CodeChallenge       string
	CodeChallengeMethod string
	RedirectURI         string
	ClientID            string
	ExpiresAt           time.Time
}

type accessToken struct {
	UserID    int64
	ClientID  string
	ExpiresAt time.Time
}

func NewService(cfg Config, users *repo.UserRepo) *Service {
	return &Service{
		cfg:    cfg,
		users:  users,
		codes:  make(map[string]authCode),
		tokens: make(map[string]accessToken),
	}
}

func (s *Service) ValidateLogin(ctx context.Context, username, password string) (int64, bool) {
	if s.users == nil {
		return 0, false
	}
	id, hash, err := s.users.GetAuthByUsername(ctx, username)
	if err != nil {
		return 0, false
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return 0, false
	}
	return id, true
}

func (s *Service) ClientID() string {
	return s.cfg.ClientID
}

func (s *Service) IssueCode(userID int64, clientID, redirectURI, codeChallenge, method string) (string, error) {
	if userID <= 0 {
		return "", fmt.Errorf("invalid user")
	}
	if strings.TrimSpace(clientID) != s.cfg.ClientID {
		return "", fmt.Errorf("invalid client_id")
	}
	if strings.TrimSpace(redirectURI) == "" {
		return "", fmt.Errorf("redirect_uri is required")
	}
	if strings.TrimSpace(codeChallenge) == "" {
		return "", fmt.Errorf("code_challenge is required")
	}
	if method == "" {
		method = "S256"
	}
	if method != "S256" {
		return "", fmt.Errorf("unsupported code_challenge_method")
	}

	code, err := randToken(24)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[code] = authCode{
		UserID:              userID,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: method,
		RedirectURI:         redirectURI,
		ClientID:            clientID,
		ExpiresAt:           time.Now().Add(s.cfg.CodeTTL),
	}
	return code, nil
}

func (s *Service) ExchangeCode(code, codeVerifier, clientID, redirectURI string) (string, int, error) {
	if clientID != s.cfg.ClientID {
		return "", 0, fmt.Errorf("invalid client_id")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.codes[code]
	if !ok {
		return "", 0, fmt.Errorf("invalid code")
	}
	delete(s.codes, code)
	if time.Now().After(data.ExpiresAt) {
		return "", 0, fmt.Errorf("code expired")
	}
	if data.RedirectURI != redirectURI {
		return "", 0, fmt.Errorf("redirect_uri mismatch")
	}
	if data.ClientID != clientID {
		return "", 0, fmt.Errorf("client_id mismatch")
	}
	if !verifyPKCE(data.CodeChallenge, data.CodeChallengeMethod, codeVerifier) {
		return "", 0, fmt.Errorf("invalid code_verifier")
	}

	token, err := randToken(32)
	if err != nil {
		return "", 0, err
	}
	ttl := int(s.cfg.AccessTokenTTL.Seconds())
	s.tokens[token] = accessToken{
		UserID:    data.UserID,
		ClientID:  clientID,
		ExpiresAt: time.Now().Add(s.cfg.AccessTokenTTL),
	}
	return token, ttl, nil
}

func (s *Service) lookupAccessToken(token string) (accessToken, bool) {
	s.mu.RLock()
	data, ok := s.tokens[token]
	s.mu.RUnlock()
	if !ok {
		return accessToken{}, false
	}
	if time.Now().After(data.ExpiresAt) {
		s.mu.Lock()
		delete(s.tokens, token)
		s.mu.Unlock()
		return accessToken{}, false
	}
	return data, true
}

func (s *Service) VerifyAccessToken(token string) bool {
	_, ok := s.lookupAccessToken(token)
	return ok
}

// UserIDFromAccessToken 校验 access token 并返回关联用户 ID。
func (s *Service) UserIDFromAccessToken(token string) (int64, bool) {
	data, ok := s.lookupAccessToken(token)
	if !ok {
		return 0, false
	}
	return data.UserID, true
}

func (s *Service) Metadata() map[string]any {
	issuer := strings.TrimRight(s.cfg.IssuerURL, "/")
	return map[string]any{
		"issuer":                                         issuer,
		"authorization_endpoint":                         issuer + "/oauth/authorize",
		"token_endpoint":                                 issuer + "/oauth/token",
		"response_types_supported":                       []string{"code"},
		"grant_types_supported":                          []string{"authorization_code"},
		"token_endpoint_auth_methods_supported":          []string{"none"},
		"code_challenge_methods_supported":               []string{"S256"},
		"scopes_supported":                               []string{"mcp"},
		"authorization_response_iss_parameter_supported": false,
	}
}

func verifyPKCE(codeChallenge, method, codeVerifier string) bool {
	if method != "S256" {
		return false
	}
	sum := sha256.Sum256([]byte(codeVerifier))
	encoded := base64.RawURLEncoding.EncodeToString(sum[:])
	return encoded == codeChallenge
}

func randToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
