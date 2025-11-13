package server

import (
	"net/http"
)

type Middleware func(http.ResponseWriter, *http.Request, http.HandlerFunc);

type Route struct {
  Pattern string;
  Middlewares []Middleware;
  Handler func(http.ResponseWriter, *http.Request);
};

