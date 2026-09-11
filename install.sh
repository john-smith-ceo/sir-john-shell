#!/usr/bin/env bash
set -euo pipefail

REPO="john-smith-ceo/sir-john-shell"
BIN="sir-john-shell"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="$HOME/.sir-john-shell"
CONFIG_FILE="$CONFIG_DIR/config"

USER_NAME="Sir"

usage() {
  echo "Usage: $0 [--user \"Name\"]"
  exit 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --user)
      [[ $# -gt 1 ]] || usage
      USER_NAME="$2"
      shift 2
      ;;
    -h|--help) usage ;;
    *) echo "Unknown option: $1"; usage ;;
  esac
done

mkdir -p "$INSTALL_DIR" "$CONFIG_DIR"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

if [[ -f "./sir-john-shell" ]]; then
  cp "./sir-john-shell" "$INSTALL_DIR/$BIN"
  echo "Installed local binary: $INSTALL_DIR/$BIN"
elif [[ -f "./dist/$BIN-$OS-$ARCH" ]]; then
  cp "./dist/$BIN-$OS-$ARCH" "$INSTALL_DIR/$BIN"
  echo "Installed local dist binary: $INSTALL_DIR/$BIN"
else
  URL="https://github.com/$REPO/releases/latest/download/$BIN-$OS-$ARCH"
  echo "Downloading $URL ..."
  curl -fsSL -o "$INSTALL_DIR/$BIN" "$URL" || { echo "Download failed"; exit 1; }
  chmod +x "$INSTALL_DIR/$BIN"
  echo "Installed: $INSTALL_DIR/$BIN"
fi

if [[ ! -f "$CONFIG_FILE" ]]; then
  printf '{"user":"%s"}\n' "$USER_NAME" > "$CONFIG_FILE"
  echo "Created $CONFIG_FILE"
else
  echo "Config already exists: $CONFIG_FILE"
fi

LAUNCHER="# Sir John Shell launcher
export SJS_USER=\"\${SJS_USER:-$USER_NAME}\"
devin() {
  if [ \$# -eq 0 ]; then
    $INSTALL_DIR/$BIN
  else
    command devin \"\$@\"
  fi
}
alias sjs=\"$INSTALL_DIR/$BIN\"
# End Sir John Shell launcher"

for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
  if [[ -f "$rc" ]]; then
    if grep -q "# Sir John Shell launcher" "$rc"; then
      sed -i.bak '/# Sir John Shell launcher/,/# End Sir John Shell launcher/d' "$rc"
      rm -f "$rc.bak"
    fi
    printf '\n%s\n' "$LAUNCHER" >> "$rc"
    echo "Updated $rc"
  fi
done

echo ""
echo "Sir John Shell installed."
echo "Run: source ~/.zshrc  (or ~/.bashrc)"
echo "Then: devin"
