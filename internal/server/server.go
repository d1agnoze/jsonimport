package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

type Server struct {
	didInit   bool // ts-go saves full initialize params but we going simple first
	documents map[string]string
	reader    *BaseReader
	writer    *BaseWriter
}

func New() Server {
	return Server{
		documents: make(map[string]string),
		reader:    NewBaseReader(os.Stdin),
		writer:    NewBaseWriter(os.Stdout),
	}
}

func (s *Server) Run() error {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			debug.PrintStack()
		}
		log.Print("shutting down server")
	}()
	log.Print("server started")

	for {
		req, err := s.read()
		if err != nil {
			if errors.Is(err, ErrInvalidRequest) {
				if invalid_err := s.sendError(nil, err); invalid_err != nil {
					panic(err)
				}
				continue
			}
			panic(err)
		}

		if !s.didInit && req.Method != methodInitialize {
			if err := s.sendError(req.ID, err); err != nil {
				panic(err)
			}
		}

		if err := s.dispatch(req); err != nil {
			panic(err)
		}
	}
}

func (s *Server) read() (*RequestMessage, error) {
	data, err := s.reader.Read()
	if err != nil {
		return nil, err
	}

	req := new(RequestMessage)
	log.Print(string(data))
	if err := json.Unmarshal(data, req); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}
	return req, nil
}

func (s *Server) sendError(id *ID, err error) error {
	code := ErrInternalError.Code

	if errCode := (*ErrorCode)(nil); errors.As(err, &errCode) {
		code = errCode.Code
	}

	return s.sendResponse(&ResponseMessage{
		ID:    id,
		Error: &ResponseError{Code: code, Message: err.Error()},
	})
}

func (s *Server) sendResponse(resp *ResponseMessage) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return s.writer.Write(data)
}

func (s *Server) dispatch(req *RequestMessage) error {
	log.Printf("%s", req)
	resp, err := s.handleMessage(req)
	if err != nil {
		if err := s.sendError(req.ID, err); err != nil {
			return err
		}
		return nil
	}
	if resp != nil {
		if err := s.sendResponse(resp); err != nil {
			if err := s.sendError(req.ID, err); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) handleMessage(req *RequestMessage) (*ResponseMessage, error) {
	switch req.Method {
	case methodInitialize:
		return s.handleInitialize(req)
	case methodInitialized:
		return s.handleInitialized(req)
	case methodTextDocumentCompletion:
		return s.handleTextDocumentCompletion(req)
	default:
		return nil, ErrInvalidRequest
	}
}
