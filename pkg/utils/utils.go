package utils

import (
  "log/slog"
  "net/http"
  "errors"
  "crypto/rand"
  "math/big"
  "encoding/json"
)

func RandomStringB64(sz uint64) string {
  const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_";
  r := make([]byte, sz);
  for i := range sz {
    n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))));
    r[i] = chars[n.Int64()]; 
  }

  return string(r);
}

/// Parses json in the provided type.
/// Returns an error message safe to send to the client
/// as well as the appropriate status code if error, 200 otherwise
/// In case of error, it is uncertain wether the T returned value is zero-valued or not
func ParseBodyJson[T any](req *http.Request) (T, error, int) {
  var data T;
  if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
    if serr, ok := err.(*json.SyntaxError); ok {
      return data, serr, http.StatusBadRequest;
    } else {
      slog.Error("Failed to parse JSON.", "error", err);
      return data, errors.New("Failed to parse JSON"), http.StatusInternalServerError;
    }
  }

  return data, nil, http.StatusOK;
}
