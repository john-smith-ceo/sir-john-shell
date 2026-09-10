package acp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"sync/atomic"
)

// Client runs devin acp as a subprocess and speaks NDJSON JSON-RPC over stdio.
type Client struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	enc     *json.Encoder
	nextID  int64
	pending map[int64]chan *RPCMessage
	notifCh chan *RPCMessage
	errCh   chan error
	wg      sync.WaitGroup
	mu      sync.Mutex
	closed  bool
}

// NewClient starts devin acp and returns a connected client.
func NewClient(ctx context.Context, devinBinary string, args []string) (*Client, error) {
	if devinBinary == "" {
		devinBinary = "devin"
	}
	cmdArgs := append([]string{"acp"}, args...)
	cmd := exec.CommandContext(ctx, devinBinary, cmdArgs...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	c := &Client{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		enc:     json.NewEncoder(stdin),
		pending: make(map[int64]chan *RPCMessage),
		notifCh: make(chan *RPCMessage, 128),
		errCh:   make(chan error, 1),
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start devin acp: %w", err)
	}

	c.wg.Add(1)
	go c.readLoop()

	return c, nil
}

// Initialize sends the ACP initialize request.
func (c *Client) Initialize(ctx context.Context) (*InitializeResponse, error) {
	req := &InitializeRequest{
		ProtocolVersion: 1,
		Capabilities:    map[string]any{},
		Info: ImplementationInfo{
			Name:    "sir-john-shell",
			Title:   "Sir John Shell",
			Version: "0.1.0",
		},
	}
	res, err := c.call(ctx, "initialize", req)
	if err != nil {
		return nil, err
	}
	var init InitializeResponse
	if err := json.Unmarshal(res.Result, &init); err != nil {
		return nil, fmt.Errorf("unmarshal initialize: %w", err)
	}
	return &init, nil
}

// NewSession sends session/new and returns the created session.
func (c *Client) NewSession(ctx context.Context, cwd string) (*NewSessionResponse, error) {
	req := &NewSessionRequest{
		Cwd:        cwd,
		McpServers: []any{},
	}
	res, err := c.call(ctx, "session/new", req)
	if err != nil {
		return nil, err
	}
	var s NewSessionResponse
	if err := json.Unmarshal(res.Result, &s); err != nil {
		return nil, fmt.Errorf("unmarshal session/new: %w", err)
	}
	return &s, nil
}

// Prompt sends a user message to the given session.
func (c *Client) Prompt(ctx context.Context, sessionID, text string) error {
	req := &PromptRequest{
		SessionID: sessionID,
		Prompt:    []PromptPart{{Type: "text", Text: text}},
	}
	_, err := c.call(ctx, "session/prompt", req)
	return err
}

// Cancel sends session/cancel.
func (c *Client) Cancel(ctx context.Context, sessionID string) error {
	req := &CancelRequest{SessionID: sessionID}
	_, err := c.call(ctx, "session/cancel", req)
	return err
}

// Notifications returns a channel of ACP notifications.
func (c *Client) Notifications() <-chan *RPCMessage {
	return c.notifCh
}

// Errors returns a channel of fatal errors.
func (c *Client) Errors() <-chan error {
	return c.errCh
}

// Close shuts down the ACP subprocess.
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

	_ = c.stdin.Close()
	c.wg.Wait()
	return c.cmd.Wait()
}

// call sends a JSON-RPC request and waits for a response.
func (c *Client) call(ctx context.Context, method string, params any) (*RPCMessage, error) {
	id := atomic.AddInt64(&c.nextID, 1)

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}

	msg := &RPCMessage{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  method,
		Params:  paramsBytes,
	}

	respCh := make(chan *RPCMessage, 1)
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, errors.New("client is closed")
	}
	c.pending[id] = respCh
	c.mu.Unlock()

	if err := c.send(msg); err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, err
	}

	select {
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, ctx.Err()
	case res := <-respCh:
		if res.Error != nil {
			return nil, fmt.Errorf("ACP error %d: %s", res.Error.Code, res.Error.Message)
		}
		return res, nil
	}
}

// send writes a single NDJSON line to devin acp stdin.
func (c *Client) send(msg *RPCMessage) error {
	return c.enc.Encode(msg)
}

// readLoop reads NDJSON from devin acp stdout and routes messages.
func (c *Client) readLoop() {
	defer c.wg.Done()
	dec := json.NewDecoder(c.stdout)
	for {
		var msg RPCMessage
		if err := dec.Decode(&msg); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrClosedPipe) {
				return
			}
			log.Printf("acp decode error: %v", err)
			select {
			case c.errCh <- err:
			default:
			}
			continue
		}

		if msg.ID != nil {
			c.handleResponse(&msg)
		} else {
			c.notifCh <- &msg
		}
	}
}

// handleResponse routes a JSON-RPC response to the waiting call.
func (c *Client) handleResponse(msg *RPCMessage) {
	id := *msg.ID
	c.mu.Lock()
	ch, ok := c.pending[id]
	delete(c.pending, id)
	c.mu.Unlock()

	if ok {
		ch <- msg
	} else {
		log.Printf("acp: unexpected response id %d", id)
	}
}
