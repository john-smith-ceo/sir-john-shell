# Sir John Shell

Browser-native UI for [Devin](https://devin.ai) with live WebSocket streaming, embedded in a single Go binary.

## Install

### One-line install

```bash
curl -fsSL https://raw.githubusercontent.com/john-smith-ceo/sir-john-shell/main/install.sh | bash
```

With your name:

```bash
curl -fsSL https://raw.githubusercontent.com/john-smith-ceo/sir-john-shell/main/install.sh | bash -s -- --user "Sir John Smith"
```

Then reload your shell and run `devin` from any project directory.

To remove:

```bash
./uninstall.sh        # keeps config
./uninstall.sh --purge # removes everything
```

### Prebuilt binaries

Download the latest release from the [Releases](https://github.com/john-smith-ceo/sir-john-shell/releases) page (e.g. [v0.1.5](https://github.com/john-smith-ceo/sir-john-shell/releases/tag/v0.1.5)):

- `sir-john-shell-darwin-arm64` — Apple Silicon (M1/M2/M3)
- `sir-john-shell-darwin-amd64` — Intel Mac
- `sir-john-shell-linux-amd64` — Linux x86_64

```bash
chmod +x sir-john-shell-darwin-arm64
mv sir-john-shell-darwin-arm64 /usr/local/bin/sir-john-shell
```

### macOS (Apple Silicon) example

```bash
mkdir -p ~/bin
cp sir-john-shell-darwin-arm64 ~/bin/sir-john-shell
chmod +x ~/bin/sir-john-shell
```

Add to `~/.zshrc`:

```zsh
devin() {
  if [ $# -eq 0 ]; then
    ~/bin/sir-john-shell
  else
    command devin "$@"
  fi
}
alias sjs='~/bin/sir-john-shell'
```

Then `source ~/.zshrc` and run `devin` from any project directory.

### Build from source

```bash
git clone https://github.com/john-smith-ceo/sir-john-shell.git
cd sir-john-shell
go build -o sir-john-shell ./cmd/sir-john-shell
```

## Run

```bash
./sir-john-shell          # TUI in terminal (default)
./sir-john-shell -mode=web   # start server and open browser
./sir-john-shell -mode=headless  # server only
```

The server listens on `127.0.0.1:8080` by default. Open `http://127.0.0.1:8080` in a browser in `web`/`headless` mode.

## CLI flags

- `-addr` — HTTP listen address (default `127.0.0.1:8080`).
- `-cwd` — working directory for the ACP session (default current directory).
- `-devin` — path to the `devin` CLI binary (default `devin`).

## Architecture

```
cmd/sir-john-shell      // entry point
internal/acp            // NDJSON/JSON-RPC subprocess client for `devin acp`
internal/server         // HTTP file server + WebSocket hub
web/static              // embedded front-end assets
web/static.go           // //go:embed wrapper
```

### ACP protocol

The backend spawns `devin acp`, speaks JSON-RPC 2.0 over NDJSON on stdin/stdout, and forwards `session/update` and custom `_cognition.ai/*` notifications to all WebSocket clients in the same session.

### WebSocket protocol

Browser connects to `/ws`. Messages:

- `welcome` — `{ type, sessionId, cwd }`.
- `session/update` — ACP update objects (`agent_message_chunk`, `agent_thought_chunk`, `usage_update`, `config_option_update`, `available_commands_update`, ...).
- `_cognition.ai/*` — agent output / lifecycle events.

Browser can send:

- `{ type: "prompt", text: "..." }`
- `{ type: "cancel" }`

## Front-end features

- Messenger-style chat (user right, agent left, 4/5 width constraint).
- Workspace file tree + file viewer with syntax highlighting.
- Live thought stream.
- Terminal / Tools / Skills switcher.
- Header dropdowns for mode and model.
- Context usage bar with model and percentage, color transitions green → yellow → orange → red.
- Focus-swap by clicking title bars (Chat, Workspace, Thoughts).
- Neo-hacker / terminal visual style.

## Development

```bash
gofmt -w ./...
go build -o sir-john-shell ./cmd/sir-john-shell
go vet ./...
```

Static assets are embedded with `//go:embed all:static`. To override the static directory during development, set `SIR_JOHN_SHELL_STATIC=./web/static`.

## License

Private project for Sir John Smith.
