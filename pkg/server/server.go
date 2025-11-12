package server

import (
  "github.com/hocman2/gogogo/pkg/server/cors"
  "github.com/hocman2/gogogo/internal/server/cors"
  "context"
  "net/http"
)

const (
  CTXServer   string = "server"
  CTXBodyJson string = "bodyjson"
);

type Middleware func(http.ResponseWriter, *http.Request, http.HandlerFunc);

type Route struct {
  pattern string;
  middlewares []Middleware;
  handler func(http.ResponseWriter, *http.Request);
};

type Server struct {
	helloMw Middleware
  mux *http.ServeMux
}

func New() *Server {
  s := &Server {
		nil,
    nil, 
  };

	s.helloMw = s.helloMidware;
  return s;
}

func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if s.mux != nil {
		s.helloMw(res, req, s.mux.ServeHTTP);
	}
}

/// Inject CORS middleware at server level for all routes
func (s *Server) WithCORS(settings *cors.CorsSettings) {
	currMw := s.helloMw;
	s.helloMw = func(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		corsW := cors_int.NewResponseWriter(w, settings);	
		currMw(corsW, req, next);
	}
}

func (s* Server) Register(routes []Route) {
  // ungly unreadable but we need to capture a different next each time
  // defined here for clarity, could be inline where it's used
  preservedNext := func(m Middleware, next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      m(w, r, next);
    });
  };

  mux := http.NewServeMux();
  for _, route := range routes {
    handler := http.HandlerFunc(route.handler);

    // reverse iterate middlewares to construct the handler chain
    for i := len(route.middlewares)-1; i >= 0; i-- {
      mw := route.middlewares[i];
      next := handler; 
      // if we dont capture next in a function call, the nested handlers end up having a reference to next instead of the previous one
      handler = preservedNext(mw, next);
    }

    mux.HandleFunc(route.pattern, handler);
  }

  s.mux = mux;
}

/// Entry middleware that sets up some server specific stuff like the response writer and context
func (s *Server) helloMidware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
  ctx := context.WithValue(r.Context(), CTXServer, s);
  r = r.WithContext(ctx);
  next.ServeHTTP(w, r); 
}
