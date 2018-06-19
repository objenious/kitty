package kitty

// Shutdown registers functions to be called when the server is stopped.
func (s *Server) Shutdown(fns ...func()) *Server {
	s.shutdown = fns
	return s
}
