package sessions

import (
  "errors"
  "time"
  "log/slog"
  "database/sql"
  "github.com/google/uuid"
  "backend/app/utils"
  "backend/app/env"
  "backend/app/models"
  "backend/app/logging"
)

type SessionId = string;
type UserId = uuid.UUID;

var (
  ErrNoUser = errors.New("No user for given session ID")
  ErrExpired = errors.New("Session expired")
)

var db *sql.DB;
var duration *time.Duration;

func Initialize(conn *sql.DB) error {
  db = conn;

  dur, err := time.ParseDuration(env.Get(env.SessionDuration)); 
  if err != nil {
    return err;
  }

  duration = &dur;
  return nil;
}

func Create(id UserId) (SessionId, *time.Time, error) {
  if db == nil {
    return "", nil, errors.New("Must call sessions.Initialize before creating/retrieving sessions");
  }

  sid := utils.RandomStringB64(128);
  expiry := time.Now().Add(*duration);
  if _, err := db.Exec("UPDATE users SET session_id = ?, session_expiry = ? WHERE id = ?", sid, expiry, id[:]); 
  err != nil {
    return "", nil, err;
  }
  return sid, &expiry, nil;
}

func GetData(id SessionId) (models.User, error) {
  if db == nil {
    return models.User {}, errors.New("Must call sessions.Initialize before creating/retrieving sessions");
  }

  row := db.QueryRow("SELECT id, email, session_expiry FROM users WHERE session_id = ?", string(id));

  var user models.User;
  var expiry time.Time;
  if err := row.Scan(&user.Id, &user.Email, &expiry); err != nil {
    if err != sql.ErrNoRows {
      slog.Error(logging.AddLocation("Error fetching user: " + err.Error()));
      return models.User{}, errors.New("Internal server error");
    } 

    return models.User{}, ErrNoUser; 
  }

  if expiry.Unix() < time.Now().Unix() {
    return models.User{}, ErrExpired;
  }

  return user, nil;
}
