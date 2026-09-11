# Sir John Shell — agent notes

## Project

- Repository path: `/home/vlad/Projects/sir-john-shell`
- Module: `sir-john-shell`
- Binary: `./sir-john-shell`
- Default listen address: `127.0.0.1:8080`

## Build

```bash
cd /home/vlad/Projects/sir-john-shell
go build -o sir-john-shell ./cmd/sir-john-shell
```

## Verify

```bash
gofmt -w ./...
go build -o sir-john-shell ./cmd/sir-john-shell
go vet ./...
```

## Run

```bash
./sir-john-shell              # TUI in terminal
./sir-john-shell -mode=web    # server + browser
./sir-john-shell -mode=headless  # server only
```

For a short command, `~/.bashrc` has a `devin()` function: `devin` starts the TUI, `devin <args>` uses the original `devin` CLI.

## Development static override

Set `SIR_JOHN_SHELL_STATIC=./web/static` to serve files from disk instead of the embedded FS.

## Important values

- User display name: `Sir John Smith`
- Agent display name: `AGENT`
- Default palette: khaki (`#9ac16a`)
- Context bar demo: 0% → 100% on every page load
- `hey-agent` must not be added to this repository

## Common pitfalls

- Port 8080 may be held by a previous binary; wait a few seconds or `killall sir-john-shell`.
- Static files are embedded; always rebuild after changing `web/static`.
