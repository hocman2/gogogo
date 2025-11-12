package server

import (
	"backend/app/auth"
	"backend/app/env"
	"backend/app/logging"
	"backend/app/models"
	"backend/app/server/cookies"
	"backend/app/sessions"
	_ "backend/app/utils"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"
)

func helperMakeNoteDir(noteId string) string {
  return env.Get(env.AssetsDir) + "/" + ASSNotesDir + "/" + noteId + "/";
}

func helperInternalServerError(w http.ResponseWriter, msg string, err error) {
  slog.Error(logging.AddLocation(msg + err.Error()));
  http.Error(w, "", http.StatusInternalServerError);
}

func preflight(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK);
}

func getHello(w http.ResponseWriter, r* http.Request) {
  w.Write([]byte("Hello"));
}

func authTokenCreate(w http.ResponseWriter, r* http.Request) {
  var body struct {
    Email string
  };

  if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
    w.WriteHeader(http.StatusBadRequest);
    w.Write([]byte("Ill formed data: " + err.Error()));
    return;
  }

  if body.Email == "" {
    w.WriteHeader(http.StatusBadRequest);
    w.Write([]byte("Missing required field: `email`"));
    return;
  }

  var uid uuid.UUID;
  server := r.Context().Value(CTXServer).(*Server);
  row := server.Db.QueryRow("SELECT id FROM users WHERE email = ?", body.Email);
  if sqlerr := row.Scan(&uid); sqlerr != nil && sqlerr != sql.ErrNoRows {
    w.WriteHeader(http.StatusInternalServerError);
    slog.Error(logging.AddLocation(sqlerr.Error()));
    return;

  } else
  if sqlerr == sql.ErrNoRows {
    uid, _ = uuid.NewRandom();

    if _, err := server.Db.Exec("INSERT INTO users (id, email) VALUES (?, ?)", uid[:], body.Email); err != nil {
      w.WriteHeader(http.StatusInternalServerError);
      slog.Error(logging.AddLocation(err.Error()));
      return;
    }
  }

  token, err := auth.CreateToken(uid);
  if err != nil {
    slog.Error(logging.AddLocation(err.Error()));
    http.Error(w, "", http.StatusInternalServerError);
  }

  if env.IsDev() {
    slog.Info("Generated auth token: ", "token", token);
  } else {
    //send email
  }

  w.WriteHeader(http.StatusOK);
}

func authTokenValidate(w http.ResponseWriter, r *http.Request) {
  token := r.FormValue("token");
  if token == "" {
    w.WriteHeader(http.StatusBadRequest);
    w.Write([]byte("Missing query parameter: `token`"));
    return;
  }

  uid, err := auth.ValidateToken(auth.Token(token));
  if err != nil {
    switch err {
      case auth.ErrTokenExpired:
        http.Error(w, "Token expired", http.StatusGone);
      case auth.ErrTokenInvalid:
        http.Error(w, "Token not found", http.StatusNotFound);
    }
    return;
  }

  sessid, sessExpiry, err := sessions.Create(uid);
  if err != nil {
    slog.Error(logging.AddLocation("Error creating session: " + err.Error()));
    http.Error(w, "", http.StatusInternalServerError);
    return;
  }

  cookie := http.Cookie {
	  Name: env.Get(env.CookieSid),
	  Value: sessid,
	  Path: "/",
	  HttpOnly: true,
	  Secure: false,
	  Expires: *sessExpiry,
	  MaxAge: 0,
	  SameSite: http.SameSiteStrictMode,
  };
  cookie = cookies.Sign(cookie);
  http.SetCookie(w, &cookie);

  w.WriteHeader(http.StatusOK);
  return;
}

func noteCreate(w http.ResponseWriter, r *http.Request) {
  noteId, _ := uuid.NewRandom();

  // we'll make only the server allowed to rwx on note dirs
  // ideally the content would be encrypted to avoid humain being able to read altogether
  noteDir := helperMakeNoteDir(noteId.String());
  if err := os.MkdirAll(noteDir, 0700); err != nil {
    slog.Error(logging.AddLocation(err.Error()));
    http.Error(w, "", http.StatusInternalServerError);
    return;
  }

  os.WriteFile(noteDir + "scripts.lua", []byte{}, 0600);
  os.WriteFile(noteDir + "metadata.json", []byte{}, 0600);
  os.WriteFile(noteDir + "content.md", []byte{}, 0600);

  titleBytes, err := io.ReadAll(r.Body);
  if err != nil {
    helperInternalServerError(w, "", err);
    return;
  }

  title := string(titleBytes);

  if title == "" {
    title = "New note";
  }

  server := r.Context().Value(CTXServer).(*Server);
  user := r.Context().Value(CTXUser).(*models.User);
  if _, err := server.Db.Exec("INSERT INTO notes (id, title, owner) VALUES (?, ?, ?)", noteId[:], title, user.Id[:]); err != nil {
    slog.Error(logging.AddLocation(err.Error()));
    http.Error(w, "", http.StatusInternalServerError);
    return;
  }

  jstr, _ := json.Marshal(map[string]string {"title": title, "id": noteId.String()});
  w.WriteHeader(http.StatusOK);
  w.Write([]byte(jstr));
}


func noteAll(w http.ResponseWriter, r* http.Request) {
  server := r.Context().Value(CTXServer).(*Server);
  user := r.Context().Value(CTXUser).(*models.User);
  rows, err := server.Db.Query("SELECT id, title FROM notes WHERE owner = ?", user.Id[:]);
  defer rows.Close();
  if err != nil {
    helperInternalServerError(w, "", err);
    return;
  }

  type NoteFetch struct {
      Id uuid.UUID `json:"id"`
      Title string `json:"title"`
  };
  var notes []NoteFetch;
  for rows.Next() {
    var noteFetch NoteFetch;
    if err := rows.Scan(&noteFetch.Id, &noteFetch.Title); err != nil {
      slog.Error(logging.AddLocation(err.Error()));
      http.Error(w, "", http.StatusInternalServerError);
      return;
    }

    notes = append(notes, noteFetch);
  }

  jstr, _ := json.Marshal(notes);
  w.WriteHeader(http.StatusOK);
  w.Write([]byte(jstr));
}

func noteGetTitle(w http.ResponseWriter, r* http.Request) {
  noteIdStr := r.PathValue("id");
  noteId, _ := uuid.Parse(noteIdStr);

  server := r.Context().Value(CTXServer).(*Server);
  row := server.Db.QueryRow("SELECT title FROM notes WHERE id = ?", noteId[:]);
  var title string;
  if err := row.Scan(&title); err != nil {
    helperInternalServerError(w, "", err);
    return;
  }

  jsn := map[string]string {"title": title};
  jstr, _ := json.Marshal(jsn);
  w.WriteHeader(http.StatusOK);
  w.Write([]byte(jstr));
}

func noteGetScript(w http.ResponseWriter, r* http.Request) {
  noteId := r.PathValue("id");
  noteDir := helperMakeNoteDir(noteId);
  http.ServeFile(w, r, noteDir + "scripts.lua");
}

func noteGetMetadata(w http.ResponseWriter, r* http.Request) {
  noteId := r.PathValue("id");
  noteDir := helperMakeNoteDir(noteId);
  http.ServeFile(w, r, noteDir + "metadata.json");
}

func noteGetContent(w http.ResponseWriter, r* http.Request) {
  noteId := r.PathValue("id");
  noteDir := helperMakeNoteDir(noteId);
  http.ServeFile(w, r, noteDir + "content.md");
}

func noteUpdateTitle(w http.ResponseWriter, r* http.Request) {
  noteIdStr := r.PathValue("id");
  noteId, _ := uuid.Parse(noteIdStr);

  newTitle, err := io.ReadAll(r.Body);
  if err != nil {
    helperInternalServerError(w, "", err);
    return;
  }

  server := r.Context().Value(CTXServer).(*Server);
  if _, err := server.Db.Exec("UPDATE notes SET title = ? WHERE id = ?", newTitle, noteId[:]); err != nil {
    helperInternalServerError(w, "", err);
    return;
  }

  w.WriteHeader(http.StatusOK);
}

func notePostScript(w http.ResponseWriter, r* http.Request) {
  noteId := r.PathValue("id");
  noteDir := helperMakeNoteDir(noteId);
  body, err := io.ReadAll(r.Body);
  if err != nil {
    helperInternalServerError(w, "", err);
    return;
  }
  os.WriteFile(noteDir + "scripts.lua", body, 0600);
  w.WriteHeader(http.StatusOK);
}

func notePostMetadata(w http.ResponseWriter, r* http.Request) {
  noteId := r.PathValue("id");
  noteDir := helperMakeNoteDir(noteId);
  body, err := io.ReadAll(r.Body);
  if err != nil {
    helperInternalServerError(w, "", err);
    return;
  }
  os.WriteFile(noteDir + "metadata.json", body, 0600);
  w.WriteHeader(http.StatusOK);
}

func notePostContent(w http.ResponseWriter, r* http.Request) {
  noteId := r.PathValue("id");
  noteDir := helperMakeNoteDir(noteId);
  body, err := io.ReadAll(r.Body);
  if err != nil {
    helperInternalServerError(w, "", err);
    return;
  }
  os.WriteFile(noteDir + "content.md", body, 0600);
  w.WriteHeader(http.StatusOK);
}
