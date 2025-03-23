package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// type definitions implementing base LSP JSON-RPC protocol

const (
	// required json-rpc version
	jsonRPCVersion = `"2.0"`

	// lsp methods
	initialize     method = "initialize"
	methodShutdown method = "shutdown"
	methodExit     method = "exit"

	textDocumentDidOpen method = "textDocument/didOpen"
)

var (
	unmarshallers = map[method]func([]byte) (any, error){}

	ErrInvalidHeader         = errors.New("lsp: invalid header")
	ErrInvalidContentLength  = errors.New("lsp: invalid content length")
	ErrNoContentLength       = errors.New("lsp: no content length")
	ErrInvalidJSONRPCVersion = errors.New("invalid JSON-RPC version")
	// Defined by JSON-RPC
	ErrParseError     = &ErrorCode{"ParseError", -32700}
	ErrInvalidRequest = &ErrorCode{"InvalidRequest", -32600}
	ErrMethodNotFound = &ErrorCode{"MethodNotFound", -32601}
	ErrInvalidParams  = &ErrorCode{"InvalidParams", -32602}
	ErrInternalError  = &ErrorCode{"InternalError", -32603}

	// Error code indicating that a server received a notification or
	// request before the server has received the `initialize` request.
	ErrServerNotInitialized = &ErrorCode{"ServerNotInitialized", -32002}
	ErrUnknownErrorCode     = &ErrorCode{"UnknownErrorCode", -32001}

	// A request failed but it was syntactically correct, e.g the
	// method name was known and the parameters were valid. The error
	// message should contain human readable information about why
	// the request failed.
	ErrRequestFailed = &ErrorCode{"RequestFailed", -32803}

	// The server cancelled the request. This error code should
	// only be used for requests that explicitly support being
	// server cancellable.
	ErrServerCancelled = &ErrorCode{"ServerCancelled", -32802}

	// The server detected that the content of a document got
	// modified outside normal conditions. A server should
	// NOT send this error code if it detects a content change
	// in it unprocessed messages. The result even computed
	// on an older state might still be useful for the client.
	//
	// If a client decides that a result is not of any use anymore
	// the client should cancel the request.
	ErrContentModified = &ErrorCode{"ContentModified", -32801}

	// The client has canceled a request and a server has detected
	// the cancel.
	ErrRequestCancelled = &ErrorCode{"RequestCancelled", -32800}
)

type method string

type DocumentUri string

type URI string

type RequestMessage struct {
	JSONRPC JSONRPCVersion `json:"jsonrpc"`
	ID      *ID            `json:"id"`
	Method  method         `json:"method"`
	Params  any            `json:"params"`
}

type ResponseMessage struct {
	JSONRPC JSONRPCVersion `json:"jsonrpc"`
	ID      *ID            `json:"id,omitempty"`
	Result  any            `json:"result"`
	Error   *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type Nullable[T any] struct {
	Value T
	Null  bool
}

type ErrorCode struct {
	Name string
	Code int32
}

type ID struct {
	str string
	int int32
}

type BaseReader struct {
	r *bufio.Reader
}

type BaseWriter struct {
	w *bufio.Writer
}

type JSONRPCVersion struct{}

func (e *ErrorCode) Error() string { return e.Name }

func (JSONRPCVersion) MarshalJSON() ([]byte, error) {
	return []byte(jsonRPCVersion), nil
}

func (*JSONRPCVersion) UnmarshalJSON(data []byte) error {
	if string(data) != jsonRPCVersion {
		return ErrInvalidJSONRPCVersion
	}
	return nil
}

func (id *ID) MarshalJSON() ([]byte, error) {
	if id.str != "" {
		return json.Marshal(id.str)
	}
	return json.Marshal(id.int)
}

func (id *ID) UnmarshalJSON(data []byte) error {
	*id = ID{}
	if len(data) > 0 && data[0] == '"' {
		return json.Unmarshal(data, &id.str)
	}
	return json.Unmarshal(data, &id.int)
}

func (r *RequestMessage) UnmarshalJSON(data []byte) error {
	var raw struct {
		JSONRPC JSONRPCVersion  `json:"jsonrpc"`
		ID      *ID             `json:"id"`
		Method  method          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}

	r.ID = raw.ID
	r.Method = raw.Method
	if r.Method == methodShutdown || r.Method == methodExit {
		// These methods have no params.
		return nil
	}

	var params any
	var err error

	if unmarshalParams, ok := unmarshallers[raw.Method]; ok {
		params, err = unmarshalParams(raw.Params)
	} else {
		// Fall back to default; it's probably an unknown message and we will probably not handle it.
		err = json.Unmarshal(raw.Params, &params)
	}

	r.Params = params

	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidRequest, err)
	}

	return nil
}

func (r *BaseReader) Read() ([]byte, error) {
	var contentLength int64

	for {
		line, err := r.r.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, io.EOF
			}
			return nil, fmt.Errorf("lsp: read header: %w", err)
		}

		if bytes.Equal(line, []byte("\r\n")) {
			break
		}

		key, value, ok := bytes.Cut(line, []byte(":"))
		if !ok {
			return nil, fmt.Errorf("%w: %q", ErrInvalidHeader, line)
		}

		if bytes.Equal(key, []byte("Content-Length")) {
			contentLength, err = strconv.ParseInt(string(bytes.TrimSpace(value)), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("%w: parse error: %w", ErrInvalidContentLength, err)
			}
			if contentLength < 0 {
				return nil, fmt.Errorf("%w: negative value %d", ErrInvalidContentLength, contentLength)
			}
		}
	}

	if contentLength <= 0 {
		return nil, ErrNoContentLength
	}

	data := make([]byte, contentLength)
	if _, err := io.ReadFull(r.r, data); err != nil {
		return nil, fmt.Errorf("lsp: read content: %w", err)
	}

	return data, nil
}

func (w *BaseWriter) Write(data []byte) error {
	if _, err := fmt.Fprintf(w.w, "Content-Length: %d\r\n\r\n", len(data)); err != nil {
		return err
	}
	if _, err := w.w.Write(data); err != nil {
		return err
	}
	return w.w.Flush()
}

func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if n.Null {
		return []byte(`null`), nil
	}
	return json.Marshal(n.Value)
}

func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	*n = Nullable[T]{}
	if string(data) == `null` {
		n.Null = true
		return nil
	}
	return json.Unmarshal(data, &n.Value)
}

func ToNullable[T any](v T) Nullable[T] {
	return Nullable[T]{Value: v}
}

func Null[T any]() Nullable[T] {
	return Nullable[T]{Null: true}
}

func NewBaseWriter(w io.Writer) *BaseWriter {
	return &BaseWriter{
		w: bufio.NewWriter(w),
	}
}

func NewBaseReader(r io.Reader) *BaseReader {
	return &BaseReader{
		r: bufio.NewReader(r),
	}
}

func unmarshallerFor[T any](data []byte) (any, error) {
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %T: %w", (*T)(nil), err)
	}
	return &v, nil
}

func assertOnlyOne(message string, values ...bool) {
	count := 0
	for _, v := range values {
		if v {
			count++
		}
	}
	if count != 1 {
		panic(message)
	}
}
