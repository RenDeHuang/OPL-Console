package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Addr                  string
	DatabaseURL           string
	PublicURL             string
	WorkspaceDomain       string
	KubeconfigPath        string
	KubeNamespace         string
	IngressClass          string
	WorkspaceImage        string
	WorkspaceStorageClass string
	SessionCookieName     string
	ConsoleUsersJSON      string
}

func Load() (Config, error) {
	cfg := Config{
		Addr:                  env("OPL_CONSOLE_ADDR", "127.0.0.1:8787"),
		DatabaseURL:           env("DATABASE_URL", "postgres://opl:secret@127.0.0.1:5432/opl_console?sslmode=disable"),
		PublicURL:             env("OPL_PUBLIC_URL", "http://127.0.0.1:8787"),
		WorkspaceDomain:       env("OPL_WORKSPACE_DOMAIN", "workspace.medopl.cn"),
		KubeconfigPath:        env("KUBECONFIG", ""),
		KubeNamespace:         env("OPL_K8S_NAMESPACE", "opl-cloud"),
		IngressClass:          env("OPL_INGRESS_CLASS", "nginx"),
		WorkspaceImage:        env("OPL_WORKSPACE_IMAGE", "ghcr.io/gaofeng21cn/one-person-lab-app:latest"),
		WorkspaceStorageClass: env("OPL_WORKSPACE_STORAGE_CLASS", "cbs"),
		SessionCookieName:     env("OPL_SESSION_COOKIE_NAME", "opl_console_session"),
		ConsoleUsersJSON:      env("OPL_CONSOLE_USERS_JSON", ""),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func env(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func EnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}
