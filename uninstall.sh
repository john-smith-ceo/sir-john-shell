#!/usr/bin/env bash
set -euo pipefail

BIN="sir-john-shell"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="$HOME/.sir-john-shell"

PURGE=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --purge) PURGE=true; shift ;;
    -h|--help)
      echo "Usage: $0 [--purge]"
      exit 0
      ;;
    *) echo "Unknown option: $1"; exit 1 ;;
  esac
done

rm -f "$INSTALL_DIR/$BIN"
echo "Removed $INSTALL_DIR/$BIN"

for rc in "$HOME/.bashrc" "$HOME/.zshrc"; do
  if [[ -f "$rc" ]]; then
    if grep -q "# Sir John Shell launcher" "$rc"; then
      sed -i.bak '/# Sir John Shell launcher/,/# End Sir John Shell launcher/d' "$rc"
      rm -f "$rc.bak"
      echo "Cleaned $rc"
    fi
  fi
done

if [[ "$PURGE" == true ]]; then
  rm -rf "$CONFIG_DIR"
  echo "Purged $CONFIG_DIR"
else
  echo "Config kept at $CONFIG_DIR (use --purge to remove)"
fi

echo "Sir John Shell uninstalled."
