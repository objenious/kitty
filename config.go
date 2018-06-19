package kitty

// Config holds configuration info for a kitty.Server.
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

// Config defines the configuration to be used by a kitty server.
func (s *Server) Config(cfg Config) *Server {
	if cfg.HTTPPort > 0 {
		s.cfg.HTTPPort = cfg.HTTPPort
	}
	if cfg.LivenessCheckPath != "" {
		s.cfg.LivenessCheckPath = cfg.LivenessCheckPath
	}
	if cfg.ReadinessCheckPath != "" {
		s.cfg.ReadinessCheckPath = cfg.ReadinessCheckPath
	}
	s.cfg.EnablePProf = cfg.EnablePProf
	return s
}

// DefaultConfig defines the default config of a kitty.Server.
var DefaultConfig = Config{
	HTTPPort:           8080,
	LivenessCheckPath:  "/alivez",
	ReadinessCheckPath: "/readyz",
	EnablePProf:        false,
}
