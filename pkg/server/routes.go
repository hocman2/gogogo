package server

var ROUTES = [...]Route {
  Route {"OPTIONS /", []Middleware{}, preflight},
  Route {"GET /hello", []Middleware{withAuth}, getHello}, 
  Route {"POST /api/auth-token-create", []Middleware{}, authTokenCreate},
  Route {"GET /api/auth-token-validate", []Middleware{}, authTokenValidate},

  Route {"GET /api/notes", []Middleware{withAuth}, noteAll},
  Route {"POST /api/note-create", []Middleware{withAuth}, noteCreate},

  Route {"GET /api/{id}/title", []Middleware{withAuth, withOwnedNote}, noteGetTitle},
  Route {"GET /api/{id}/scripts", []Middleware{withAuth, withOwnedNote}, noteGetScript},
  Route {"GET /api/{id}/metadata", []Middleware{withAuth, withOwnedNote}, noteGetMetadata},
  Route {"GET /api/{id}/content", []Middleware{withAuth, withOwnedNote}, noteGetContent},

  Route {"POST /api/{id}/title", []Middleware{withAuth, withOwnedNote}, noteUpdateTitle},
  Route {"POST /api/{id}/scripts", []Middleware{withAuth, withOwnedNote}, notePostScript},
  Route {"POST /api/{id}/metadata", []Middleware{withAuth, withOwnedNote}, notePostMetadata},
  Route {"POST /api/{id}/content", []Middleware{withAuth, withOwnedNote}, notePostContent},
}
