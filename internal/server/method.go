package server

import "d1agnoze/jsonimport/internal/u"

func (s *Server) handleInitialize(req *RequestMessage) (*ResponseMessage, error) {
	resp := js{
		"capabilities": js{
			"textDocumentSync": float64(1), // Full sync
			"completionProvider": js{
				"triggerCharacters": []string{"."},
			},
		},
	}

	msg := ResponseMessage{ID: req.ID, Result: resp}
	return &msg, nil
}

// just a notification, no response needed
func (*Server) handleInitialized(*RequestMessage) (*ResponseMessage, error) {
	return nil, nil
}

func (s *Server) handleTextDocumentCompletion(req *RequestMessage) (*ResponseMessage, error) {
	res := completionList{
		IsIncomplete: false,
		Items: []completionItem{
			{
				Label:      "Foo",
				Kind:       completionKindField,
				Preselect:  u.Pt(true),
				InsertText: u.Pt("Bar"),
			},
			{
				Label:      "John",
				Kind:       completionKindField,
				InsertText: u.Pt("Doe"),
			},
		},
	}
	resp := ResponseMessage{ID: req.ID, Result: res}
	return &resp, nil
}
