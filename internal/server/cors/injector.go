package cors_int

import (
	"net"
	"bufio"
	"net/http"
	"github.com/hocman2/gogogo/pkg/server/cors"
);


/// This is the response writer injected in the server when requested
type ResponseWriter struct {
  w http.ResponseWriter
	corsSettings *cors.CorsSettings
  corsHeadersSet bool
}

func NewResponseWriter(w http.ResponseWriter, settings *cors.CorsSettings) *ResponseWriter {
  return &ResponseWriter {
    w,
		settings,
    false,
  }; 
}

func (w *ResponseWriter) SetHeaders() {
  w.Header().Set(
		"Access-Control-Allow-Origin", 
		w.corsSettings.AllowedOrigins(),
	);

  w.Header().Set(
		"Access-Control-Allow-Headers",
		w.corsSettings.AllowedHeaders(),
	);

  w.Header().Set(
		"Access-Control-Allow-Methods",
		w.corsSettings.AllowedMethods(),
	);

  w.Header().Set(
		"Access-Control-Expose-Headers",
		w.corsSettings.ExposedHeaders(),
	);

  w.Header().Set(
		"Access-Control-Allow-Credentials",
		w.corsSettings.AllowCredentials(),
	);

  w.corsHeadersSet = true;
}

func (w *ResponseWriter) Header() http.Header {
  return w.w.Header();
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
  if !w.corsHeadersSet {
    w.SetHeaders();
  }

  w.w.WriteHeader(statusCode);
}

func (w *ResponseWriter) Write(b []byte) (int, error) {
  if !w.corsHeadersSet {
    w.SetHeaders();
  }

  return w.w.Write(b);
}

func (w *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.w.(http.Hijacker); ok {
		return h.Hijack();
	}

	return nil, nil, http.ErrNotSupported;
}

func (w *ResponseWriter) Flush() {
if f, ok := w.w.(http.Flusher); ok {
		f.Flush();
	}
}

func (w *ResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := w.w.(http.Pusher); ok {
		return p.Push(target, opts);
	}

	return http.ErrNotSupported;
}
