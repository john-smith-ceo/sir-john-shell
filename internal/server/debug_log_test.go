package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestThemeLogEndpointWritesJSONL(t *testing.T) {
	logDir := t.TempDir()
	s := &Server{logDir: logDir}
	req := httptest.NewRequest(http.MethodPost, "/debug/theme-log", strings.NewReader(`{"event":"polygon_ready","data":{"version":1}}`))
	rec := httptest.NewRecorder()
	s.handleThemeLog(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	data, err := os.ReadFile(filepath.Join(logDir, "theme-flow.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var entry map[string]any
	if err := json.Unmarshal(data[:len(data)-1], &entry); err != nil {
		t.Fatal(err)
	}
	if entry["event"] != "polygon_ready" {
		t.Fatalf("event = %v", entry["event"])
	}
}

func TestSafePathAllowsWorkspacePathsAndBlocksTraversal(t *testing.T) {
	if got, err := safePath("/workspace", "internal/acp"); err != nil || got != "/workspace/internal/acp" {
		t.Fatalf("safe workspace path = %q, %v", got, err)
	}
	for _, name := range []string{"../outside", "a/../../outside", "/etc/passwd"} {
		if _, err := safePath("/workspace", name); err == nil {
			t.Fatalf("safePath(%q) allowed traversal", name)
		}
	}
}
