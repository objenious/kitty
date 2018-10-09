package kitty

import kithttp "github.com/go-kit/kit/transport/http"

// Config holds configuration info for kitty.HTTPTransport.
type Config struct {
	// LivenessCheckPath is the path of the health handler (default: "/alivez").
	LivenessCheckPath string
	// ReadinessCheckPath is the path of the readiness handler (default: "/readyz").
	ReadinessCheckPath string
	// HTTPPort is the port the server will listen on (default: 8080).
	HTTPPort int
	// EnablePProf enables pprof urls (default: false).
	EnablePProf bool
	// EncodeResponse defines the default response encoder for all endpoints (by default: EncodeJSONResponse). It can be overriden for a specific endpoint.
	EncodeResponse kithttp.EncodeResponseFunc
}

// DefaultConfig defines the default config of kitty.HTTPTransport.
var DefaultConfig = Config{
	HTTPPort:           8080,
	LivenessCheckPath:  "/alivez",
	ReadinessCheckPath: "/readyz",
	EnablePProf:        false,
	EncodeResponse:     kithttp.EncodeJSONResponse,
}
