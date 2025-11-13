package cors

import (
  "strings"
	srv "github.com/hocman2/gogogo/pkg/server/definitions"
	"net/http"
);

// ---------------------------------

var ALLOWED_HEADERS_DEFAULT = []string {
  "Content-Type",
  "Origin",
  "Accept",
  "Authorization",
  "X-Requested-With",
};

var ALLOWED_METHODS_ALL = []string {
  "GET",
  "POST",
  "OPTIONS",
  "PUT",
  "PATCH",
  "DELETE",
};

var EXPOSED_HEADERS_DEFAULT = []string {
  "Content-Length",
  "Content-Type",
  "Date",
  "X-Requested-With",
"Authorization",
};

const ALLOW_CREDENTIALS_DEFAULT = true;

// ---------------------------------

type CorsSettings struct {
	allowedOrigins []string
	allowedHeaders []string
	allowedMethods []string
	exposedHeaders []string
	allowCredentials bool
}

func NewDefault(allowedOrigins []string) *CorsSettings {
	return &CorsSettings {
		allowedOrigins,
		ALLOWED_HEADERS_DEFAULT,
		ALLOWED_METHODS_ALL,
		EXPOSED_HEADERS_DEFAULT,
		ALLOW_CREDENTIALS_DEFAULT,
	};
}

func NewComplete(
	allowedOrigins []string,
	allowedHeaders []string,
	allowedMethods []string,
	exposedHeaders []string,
	allowCredentials bool,
) *CorsSettings {
	return &CorsSettings {
		allowedOrigins,
		allowedHeaders,
		allowedMethods,
		exposedHeaders,
		allowCredentials,
	};
}


func (c *CorsSettings) AllowedOrigins() string {
  return strings.Join(c.allowedOrigins, ", ");
}
func (c *CorsSettings) AllowedHeaders() string {
  return strings.Join(c.allowedHeaders, ", ");
}
func (c *CorsSettings) AllowedMethods() string {
  return strings.Join(c.allowedMethods, ", ");
}
func (c *CorsSettings) ExposedHeaders() string {
  return strings.Join(c.exposedHeaders, ", ");
}
func (c *CorsSettings) AllowCredentials() string {
  if c.allowCredentials {
    return "true";
  } else {
    return "false";
  }
}

func PreflightHandler() srv.Route {
	return srv.Route {
		"OPTIONS /",
		[]srv.Middleware {},
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK);
		},
	};
}
