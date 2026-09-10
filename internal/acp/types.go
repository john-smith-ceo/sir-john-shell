package acp

import (
	"encoding/json"
)

// RPCMessage is a generic JSON-RPC envelope.
type RPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int64          `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError is a JSON-RPC error object.
type RPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// ImplementationInfo describes a client or agent implementation.
type ImplementationInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Version string `json:"version"`
}

// AuthMethod describes an available authentication method.
type AuthMethod struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// InitializeRequest is the ACP initialize method request.
type InitializeRequest struct {
	ProtocolVersion int                `json:"protocolVersion"`
	Capabilities    map[string]any     `json:"capabilities"`
	Info            ImplementationInfo `json:"info"`
}

// InitializeResponse is the ACP initialize method response.
type InitializeResponse struct {
	ProtocolVersion   int                `json:"protocolVersion"`
	AgentCapabilities map[string]any     `json:"agentCapabilities"`
	AgentInfo         ImplementationInfo `json:"agentInfo"`
	AuthMethods       []AuthMethod       `json:"authMethods"`
}

// NewSessionRequest is the ACP session/new request.
type NewSessionRequest struct {
	Cwd        string `json:"cwd"`
	McpServers []any  `json:"mcpServers"`
}

// ConfigOption is a session configuration option returned by session/new.
type ConfigOption struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Category     string   `json:"category"`
	Type         string   `json:"type"`
	CurrentValue string   `json:"currentValue"`
	Options      []Option `json:"options"`
}

// Option is a single option value.
type Option struct {
	Value string `json:"value"`
	Name  string `json:"name"`
}

// Modes describes the available session modes.
type Modes struct {
	CurrentModeID string `json:"currentModeId"`
}

// NewSessionResponse is the ACP session/new response.
type NewSessionResponse struct {
	SessionID     string         `json:"sessionId"`
	Modes         Modes          `json:"modes"`
	ConfigOptions []ConfigOption `json:"configOptions"`
}

// PromptPart is one part of a user prompt.
type PromptPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// PromptRequest is the ACP session/prompt request.
type PromptRequest struct {
	SessionID string       `json:"sessionId"`
	Prompt    []PromptPart `json:"prompt"`
}

// CancelRequest is the ACP session/cancel request.
type CancelRequest struct {
	SessionID string `json:"sessionId"`
}
