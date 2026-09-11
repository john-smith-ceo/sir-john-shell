package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

var (
	styleUser    = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffbd2e"))
	styleAgent   = lipgloss.NewStyle().Foreground(lipgloss.Color("#9ac16a"))
	styleThought = lipgloss.NewStyle().Foreground(lipgloss.Color("#8be9fd"))
	styleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
	styleErr     = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5f56"))
)

type message struct {
	kind string
	text string
}

// Run starts the terminal UI connected to the local WebSocket server.
func Run(wsAddr string, ctx context.Context) error {
	u := url.URL{Scheme: "ws", Host: wsAddr, Path: "/ws"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("dial %s: %w", u.String(), err)
	}
	defer c.Close()

	p := tea.NewProgram(initialModel(c), tea.WithAltScreen(), tea.WithMouseCellMotion())

	go func() {
		for {
			_, data, err := c.ReadMessage()
			if err != nil {
				p.Send(errMsg{err})
				return
			}
			p.Send(rawMsg(data))
		}
	}()

	go func() {
		<-ctx.Done()
		c.WriteJSON(map[string]string{"type": "cancel"})
		c.Close()
		p.Quit()
	}()

	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

type errMsg struct{ err error }
type rawMsg []byte

type model struct {
	conn     *websocket.Conn
	viewport viewport.Model
	input    textinput.Model
	messages []message
	ready    bool
	width    int
	height   int
}

func initialModel(conn *websocket.Conn) model {
	ti := textinput.New()
	ti.Placeholder = "Введите сообщение..."
	ti.Focus()
	ti.CharLimit = 8192
	ti.Width = 60

	vp := viewport.New(80, 20)
	vp.SetContent(``)

	return model{
		conn:     conn,
		input:    ti,
		viewport: vp,
		messages: []message{{kind: "system", text: "Connecting..."}},
	}
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = max(10, msg.Width-4)
		m.viewport.Width = max(10, msg.Width-4)
		m.viewport.Height = max(5, msg.Height-4)
		m.ready = true
		m.updateViewport()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.conn.WriteJSON(map[string]string{"type": "cancel"})
			return m, tea.Quit
		case tea.KeyEnter:
			text := strings.TrimSpace(m.input.Value())
			if text == "" {
				return m, nil
			}
			if text == "/quit" || text == "/exit" {
				return m, tea.Quit
			}
			if text == "/cancel" {
				m.conn.WriteJSON(map[string]string{"type": "cancel"})
				m.append("system", "[cancel]")
				m.input.SetValue("")
				return m, nil
			}
			m.append("user", text)
			if err := m.conn.WriteJSON(map[string]string{"type": "prompt", "text": text}); err != nil {
				log.Printf("send prompt: %v", err)
			}
			m.input.SetValue("")
		}

	case rawMsg:
		m.handleMsg([]byte(msg))

	case errMsg:
		m.append("err", fmt.Sprintf("[connection error] %v", msg.err))
		m.updateViewport()
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}
	header := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ac16a")).Render("SIR JOHN SHELL // CLI")
	return fmt.Sprintf("%s\n%s\n\n%s", header, m.viewport.View(), m.input.View())
}

func (m *model) append(kind, text string) {
	if text == "" {
		return
	}
	m.messages = append(m.messages, message{kind: kind, text: text})
}

func (m *model) appendChunk(kind, text string) {
	if text == "" {
		return
	}
	if n := len(m.messages); n > 0 && m.messages[n-1].kind == kind {
		m.messages[n-1].text += text
		return
	}
	m.append(kind, text)
}

func (m *model) updateViewport() {
	lines := make([]string, 0, len(m.messages))
	for _, msg := range m.messages {
		switch msg.kind {
		case "user":
			lines = append(lines, styleUser.Render("> ")+styleUser.Render(msg.text))
		case "agent":
			lines = append(lines, styleAgent.Render(msg.text))
		case "thought":
			lines = append(lines, styleThought.Render(msg.text))
		case "err":
			lines = append(lines, styleErr.Render(msg.text))
		default:
			lines = append(lines, styleDim.Render(msg.text))
		}
	}
	m.viewport.SetContent(strings.Join(lines, "\n"))
	m.viewport.GotoBottom()
}

func (m *model) handleMsg(data []byte) {
	var msg map[string]any
	if err := json.Unmarshal(data, &msg); err != nil {
		m.append("err", string(data))
		m.updateViewport()
		return
	}

	if t, _ := msg["type"].(string); t == "welcome" {
		sid, _ := msg["sessionId"].(string)
		cwd, _ := msg["cwd"].(string)
		m.append("system", fmt.Sprintf("[session %s @ %s]", sid, cwd))
		m.updateViewport()
		return
	}

	method, _ := msg["method"].(string)
	params, _ := msg["params"].(map[string]any)

	switch method {
	case "session/update":
		up, _ := params["update"].(map[string]any)
		if up == nil {
			return
		}
		typ, _ := up["sessionUpdate"].(string)
		switch typ {
		case "agent_message_chunk":
			m.appendChunk("agent", chunkText(up["content"]))
		case "agent_thought_chunk":
			m.appendChunk("thought", chunkText(up["content"]))
		case "usage_update":
			used, _ := up["used"].(float64)
			size, _ := up["size"].(float64)
			if size > 0 {
				pct := int(used / size * 100)
				m.append("system", fmt.Sprintf("[context %d%%]", pct))
			}
		case "session_info_update":
			if title, ok := up["title"].(string); ok {
				m.append("system", fmt.Sprintf("[title: %s]", title))
			}
		case "current_mode_update":
			if id, ok := up["currentModeId"].(string); ok {
				m.append("system", fmt.Sprintf("[mode: %s]", id))
			}
		case "config_option_update":
			m.append("system", "[config updated]")
		}
	case "_cognition.ai/output":
		ch, _ := params["channel"].(string)
		lv, _ := params["level"].(string)
		text, _ := params["message"].(string)
		m.append("system", fmt.Sprintf("[%s/%s] %s", ch, lv, text))
	case "_cognition.ai/agent_stopped":
		m.append("system", "[done]")
	case "_cognition.ai/thinking_complete":
	}
	m.updateViewport()
}

func chunkText(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	m, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	if t, ok := m["type"].(string); ok && t == "text" {
		if s, ok := m["text"].(string); ok {
			return s
		}
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
