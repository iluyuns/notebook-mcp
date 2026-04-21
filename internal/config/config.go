package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                 string
	PostgresDSN          string
	MCPPath              string
	DefaultPerPage       int
	OAuthIssuerURL       string
	OAuthClientID        string
	InitialInviteCode    string
	InitialInviteMaxUses int
	AccessTokenTTL       int
	CodeTTL              int
}

func Load() Config {
	return Config{
		Port:                 getEnv("PORT", "8088"),
		PostgresDSN:          getEnv("POSTGRES_DSN", ""),
		MCPPath:              getEnv("MCP_PATH", "/mcp"),
		DefaultPerPage:       getEnvInt("DEFAULT_QUERY_LIMIT", 20),
		OAuthIssuerURL:       getEnv("OAUTH_ISSUER_URL", "http://localhost:8088"),
		OAuthClientID:        getEnv("OAUTH_CLIENT_ID", "cursor-private-notebook"),
		InitialInviteCode:    getEnv("INITIAL_INVITE_CODE", ""),
		InitialInviteMaxUses: getEnvInt("INITIAL_INVITE_MAX_USES", 1000),
		AccessTokenTTL:       getEnvInt("OAUTH_ACCESS_TOKEN_TTL_SECONDS", 3600),
		CodeTTL:              getEnvInt("OAUTH_CODE_TTL_SECONDS", 300),
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}
