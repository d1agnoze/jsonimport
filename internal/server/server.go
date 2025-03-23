package server

import "net/rpc"

type Server struct {
	documents map[string]string
}

func New() Server {
	return Server{make(map[string]string)}
}

func (s *Server) Initialize(args js, rep *js) error {
	*rep = js{
		"capabilities": js{
			"textDocumentSync": float64(1), // Full sync
		},
	}
	return nil
}

func (s *Server) Dispatch(args map[string]interface{}, reply *interface{}) error {
        return nil
}


func (s *Server) DispatchV2(method method, args map[string]any, reply *js) error {
	switch method {
	case "initialize":
		return s.Initialize(args, reply)
	case textDocumentDidOpen:
		return s.DidOpen(args, reply)
	default:
		return rpc.ErrShutdown // Unknown method
	}
}
