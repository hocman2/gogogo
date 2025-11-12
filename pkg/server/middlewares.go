package server

import (
	"backend/app/env"
	"backend/app/logging"
	"backend/app/models"
	"backend/app/server/cookies"
	"backend/app/sessions"
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

func withAuth(res http.ResponseWriter, req* http.Request, next http.HandlerFunc) {
	signedCookie, err := req.Cookie(env.Get(env.CookieSid));
	if err != nil {
		http.Error(res, "", http.StatusUnauthorized);
		return;
	}

	sid, err := cookies.Verify(signedCookie);
	if err != nil {
		http.Error(res, "", http.StatusUnauthorized);
		return;
	}

	user, err := sessions.GetData(sid.Value);
	if err != nil {
		switch err {
			case sessions.ErrNoUser:
			http.Error(res, "Invalid session id !", http.StatusUnauthorized);
			return;
			case sessions.ErrExpired:
			res.Header()["WWW-Authenticate"] = []string {"realm /api/auth-token-create"};
			http.Error(res, "Session expired", http.StatusUnauthorized)
			return;
			default:
			http.Error(res, "", http.StatusInternalServerError);
			return;
		}
	}

	// update request ocntext with the user
	ctx := context.WithValue(req.Context(), CTXUser, &user);
	req = req.WithContext(ctx);

	next(res, req);
}

func withOwnedNote(res http.ResponseWriter, req* http.Request, next http.HandlerFunc) {
  noteIdStr := req.PathValue("id");
  if noteIdStr == "" {
    http.Error(res, "Missing id wildcard parameter", http.StatusBadRequest);
    return;
  }

  noteId, err := uuid.Parse(noteIdStr);
  if err != nil {
    http.Error(res, "Bad id format", http.StatusBadRequest);
    return;
  }

  server := req.Context().Value(CTXServer).(*Server);
  user := req.Context().Value(CTXUser).(*models.User);
  if user == nil {
    log.Fatal(logging.AddLocation("withOwnedNote middleware called but there was no user. Make sure to call the withAuth middleware first"));
    http.Error(res, "", http.StatusInternalServerError);
    return;
  }

  row := server.Db.QueryRow("SELECT owner FROM notes WHERE id = ?", noteId[:]);
  var noteOwner uuid.UUID;
  if err := row.Scan(&noteOwner); err != nil {
    if err == sql.ErrNoRows {
      http.Error(res, "", http.StatusUnauthorized);
      return;
    }

    slog.Error(logging.AddLocation(err.Error()));
    http.Error(res, "", http.StatusInternalServerError);
    return;
  }

  if noteOwner.String() != user.Id.String() {
    http.Error(res, "", http.StatusUnauthorized);
    return;
  }

  next(res, req);
}
