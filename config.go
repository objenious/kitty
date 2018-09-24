package kitty

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
}

// DefaultConfig defines the default config of kitty.HTTPTransport.
var DefaultConfig = Config{
	HTTPPort:           8080,
	LivenessCheckPath:  "/alivez",
	ReadinessCheckPath: "/readyz",
	EnablePProf:        false,
}
