#!/bin/sh
set -eu

REPO="dittofleet/navi"
DEST="${NAVI_INSTALL_DIR:-$HOME/.local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
  arm64|aarch64) ARCH=arm64 ;;
  x86_64) ARCH=x64 ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

ASSET="navi-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

mkdir -p "$DEST"
# Staged inside the destination so the install is a same-filesystem
# rename. Downloading to $TMPDIR would make the final step a copy over the
# live binary, which an interruption could leave truncated.
TMP=$(mktemp "$DEST/.navi.XXXXXX")
trap 'rm -f "$TMP"' EXIT

echo "Downloading $URL..." >&2
curl -fsSL "$URL" -o "$TMP"
chmod 755 "$TMP"
mv "$TMP" "$DEST/navi"
echo "Installed navi to $DEST/navi" >&2

# NAVI_TOPIC (with NAVI_SERVER and NAVI_TOKEN if needed) in the
# environment makes a fresh remote machine a single curl-pipe: setup saves
# whatever they hold. Without them an existing config, perhaps synced from
# another machine, is left as it is.
CONFIG_FILE="${XDG_CONFIG_HOME:-$HOME/.config}/navi/config.json"
if [ -n "${NAVI_TOPIC:-}" ]; then
  "$DEST/navi" setup >/dev/null
  echo "Saved the topic from NAVI_TOPIC to $CONFIG_FILE" >&2
elif [ -f "$CONFIG_FILE" ]; then
  echo "Using the existing config at $CONFIG_FILE" >&2
else
  echo "Run \`navi setup\` to pick a topic and see how to subscribe on your phone." >&2
fi

case ":$PATH:" in
  *":$DEST:"*) ;;
  *) echo "Note: $DEST is not in \$PATH. Add it to your shell profile to use navi." >&2 ;;
esac
