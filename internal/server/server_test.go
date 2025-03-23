package server

import "testing"

func Test_server_Initialize(t *testing.T) {
	server := New()
	args := js{
		"processId": float64(1234),
		"rootUri":   "file:///path/to/project",
	}

	var reply js

	err := server.Initialize(args, &reply)
	if err != nil {
		t.Fatalf("server.Initialize failed: %s", err)
	}

	capabilities, ok := reply["capabilities"].(js)
	if !ok {
		t.Fatalf("expected capabilities in reply")
	}

	syncKind, ok := capabilities["textDocumentSync"].(float64)
	if !ok || syncKind != 1 {
		t.Fatalf("expected textDocumentSync = 1, got %#v", syncKind)
	}
}

func Test_server_DidOpen(t *testing.T) {
	server := New()

	// Simulate didOpen notification
	args := map[string]any{
		"textDocument": map[string]any{
			"uri":  "file:///test.go",
			"text": "package main",
		},
	}
	var reply js

	// Call the method
	err := server.DidOpen(args, &reply)
	if err != nil {
		t.Fatalf("DidOpen failed: %v", err)
	}

	// Check session data
	text, exists := server.documents["file:///test.go"]
	if !exists || text != "package main" {
		t.Errorf("Expected document 'file:///test.go' with text 'package main', got %v", text)
	}

	// Notifications don’t need a reply
	if reply != nil {
		t.Errorf("Expected nil reply, got %v", reply)
	}
}

func Test_server_Dispatch(t *testing.T) {
	server := New()
	// Simulate JSON-RPC call
	args := js{"processId": float64(1234), "rootUri": "file:///path/to/project"}

	t.Run("test server initialize", func(t *testing.T) {
		var reply js
		if err := server.DispatchV2("initialize", args, &reply); err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}

		capabilities, ok := reply["capabilities"].(js)
		if !ok || capabilities["textDocumentSync"] != float64(1) {
			t.Errorf("Expected textDocumentSync = 1, got %v", capabilities["textDocumentSync"])
		}
	})

}
