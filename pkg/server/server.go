package server

import (
  "github.com/hocman2/gogogo/pkg/server/cors"
  defs "github.com/hocman2/gogogo/pkg/server/definitions"
  "github.com/hocman2/gogogo/internal/server/cors"
  "context"
  "net/http"
	"reflect"
)

const (
  CTXServer   string = "server"
  CTXBodyJson string = "bodyjson"
);

type UnregisteredServer struct {
	helloMw defs.Middleware
	autoRoutes []defs.Route
	ctxValues map[any]any
}

type Server struct {
	helloMw defs.Middleware
  mux *http.ServeMux
}

func New() *UnregisteredServer {
	s := &UnregisteredServer {
		nil,
    make([]defs.Route, 0), 
		make(map[any]any),
  }

	return s
}

/// Inject CORS middleware at server level for all routes
/// This function also automatically adds a preflight route at schema "OPTIONS /"
func (s *UnregisteredServer) WithCORS(settings *cors.CorsSettings) *UnregisteredServer {
	currMw := s.helloMw;
	s.helloMw = func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		corsW := cors_int.NewResponseWriter(w, settings)
		if currMw != nil {
			currMw(corsW, req, next)
		}
	}

	s.autoRoutes = append(
		s.autoRoutes,
		cors.PreflightHandler(),
		);

	return s;
}

/// A value that will be attached to every request context
func (s* UnregisteredServer) WithValue(key, value any) *UnregisteredServer {
	if key == nil {
		panic("Key must not be nil");
	}
	if !reflect.TypeOf(key).Comparable() {
		panic("Key must be of a comparable type");
	}

	s.ctxValues[key] = value;
	return s;
}

func (s* UnregisteredServer) Register(routes []defs.Route) *Server {
  // ungly unreadable but we need to capture a different next each time
  // defined here for clarity, could be inline where it's used
  preservedNext := func(m defs.Middleware, next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      m(w, r, next);
    });
  };

  mux := http.NewServeMux();
	for _, route := range s.autoRoutes {
		handler := http.HandlerFunc(route.Handler);
		mux.HandleFunc(route.Pattern, handler);
	}

  for _, route := range routes {
    handler := http.HandlerFunc(route.Handler);

    // reverse iterate middlewares to construct the handler chain
    for i := len(route.Middlewares)-1; i >= 0; i-- {
      mw := route.Middlewares[i];
      next := handler; 
      // if we dont capture next in a function call, the nested handlers end up having a reference to next instead of the previous one
      handler = preservedNext(mw, next);
    }

    mux.HandleFunc(route.Pattern, handler);
  }

	srv := &Server {
		s.helloMw,
		mux,
	}

	initialMidware := func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		ctx := context.WithValue(r.Context(), CTXServer, srv)
		for k, v := range s.ctxValues {
			ctx = context.WithValue(ctx, k, v)
		}
		r = r.WithContext(ctx)
		next(w, r)
	}

	helloMw := srv.helloMw
	srv.helloMw = func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if helloMw != nil {
			initialMidware(w, r, func(w http.ResponseWriter, r *http.Request) {
				helloMw(w, r, next)
			})
		} else {
			initialMidware(w, r, next)
		}
	}

	return srv
}

func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if s.mux != nil {
		s.helloMw(res, req, s.mux.ServeHTTP);
	}
}

type Route = defs.Route;
type Middleware = defs.Middleware;
