package server

func (s *Server) DidOpen(args map[string]any, reply *js) error {
	textDoc := args["textDocument"].(js)
	uri := textDoc["uri"].(string)
	text := textDoc["text"].(string)
	s.documents[uri] = text
	*reply = nil // No response for notifications
	return nil
}
