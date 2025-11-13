package env

import (
  "log"
  "log/slog"
  "os"
  "github.com/joho/godotenv"
);

type EnvKey string

const (
  Environment          EnvKey = "ENV"
	DBUrl								 EnvKey = "DATABASE_URL"
  ClientUrl            EnvKey = "CLIENT_URL"
  CookieSecret         EnvKey = "COOKIE_SECRET"
  CookieSid            EnvKey = "COOKIE_SID"
  AssetsDir            EnvKey = "ASSETS_DIR"
  AuthTokenDuration    EnvKey = "AUTH_TOKEN_DURATION"
  SessionDuration      EnvKey = "SESSION_DURATION"
);

type Env struct {
  environment string;
  devenv bool;
}

var innerenv *Env;

func NewEnv() {
  if err := godotenv.Load(); err != nil {
    log.Fatal("Error loading .env file");
  }

  environment := Get(Environment);
  if len(environment) == 0 {
    slog.Warn("ENV environment variable not found, assuming production environment");
    environment = "production";
  } 

  devenv := false;
  if environment == "development" {
    devenv = true;
  }

  innerenv = &Env {
    environment: environment,
    devenv: devenv,
  };
}

func (env* Env) Environment() string {
  return env.environment;
}

/// Shorthand for os.Getenv with predefined env keys
func Get(key EnvKey) string {
  return os.Getenv(string(key));
}

func IsDev() bool {
  if innerenv == nil {
    panic("Attempted to access env but NewEnv() wasn't called prior");
  }

  return innerenv.devenv;
}
