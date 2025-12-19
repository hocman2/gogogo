// Example server runtime using cors, rng and crypto features
// This cmd is purposefuly overly versatile to support many use cases 
package main

import (
  "log"
  "log/slog"
  "net/http"
  "crypto/rand"
  "math/big"
  "strconv"

  "github.com/hocman2/gogogo/pkg/server"
	"github.com/hocman2/gogogo/pkg/server/cors"
  "github.com/hocman2/gogogo/pkg/auth"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug);
	slog.Debug("🤔 Can you see this ? Debug logs are enabled for dev environments.");

  // making sure the environment allows crypto rng
  if _, err := rand.Int(rand.Reader, big.NewInt(10)); err != nil {
    log.Fatal("❌ Failed crypto-secure RNG: " + err.Error());
  }
  slog.Info("✅ Passed test: crypto RNG");

  if err := auth.InitializeTokenGenerator("15m"); err != nil {
    log.Fatal("❌ Failed to initialize token generator: ", err.Error());
  }
  slog.Info("✅ Auth: token generator initialized");

	corsSettings := cors.NewDefault(
		[]string {},
	);

	server := server.
		New().
		WithCORS(corsSettings).
		Register([]server.Route {});

	const PORT = 9000;
	slog.Info("⚙️ Server: started", "port", PORT);
	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(PORT), server));
}
