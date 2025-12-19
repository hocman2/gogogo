package auth

import (
  "errors"
  "log/slog"
  "time"
  "github.com/hocman2/gogogo/pkg/utils"
  "github.com/google/uuid"
)

type Token string;
type AuthToken struct {
  userId uuid.UUID
  expiry int64
}

var (
  ErrTokenExpired = errors.New("Token expired")
  ErrTokenInvalid = errors.New("Token not found")
);

var authTokens map[Token]AuthToken = make(map[Token]AuthToken);
var tokenDuration *time.Duration;

func InitializeTokenGenerator(tokenDurationStr string) error {
  if tokenDuration == nil {
    dur, err := time.ParseDuration(tokenDurationStr);
    if err != nil {
      return err;
    }
    tokenDuration = &dur;
  } else {
    slog.Warn("Auth token generator was initialized twice");
    // would be neat to have a stack trace here
  }

  return nil;
}

func CreateToken(uid uuid.UUID) (Token, error) {
  if tokenDuration == nil {
    return Token(""), errors.New("Must call InitializeTokenGenerator before any token creation !");
  }

  token := Token(utils.RandomStringB64(128));
  expiry := time.Now().Add(*tokenDuration);
  authTokens[token] = AuthToken { uid, expiry.Unix() };
  return token, nil;
}

func ValidateToken(token Token) (uuid.UUID, error) {
  tokendata, ok := authTokens[token];
  if !ok {
    return [16]byte{0}, ErrTokenInvalid;
  }
  
  delete(authTokens, token);

  if time.Now().Unix() > tokendata.expiry {
    return [16]byte{0}, ErrTokenExpired;
  }

  return tokendata.userId, nil;
}
