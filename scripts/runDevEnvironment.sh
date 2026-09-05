#!/bin/bash
# Launches the full local dev environment for the Ships game in one go:
#   0. Shared config (files/.env, files/ssl/*) - ensure it's wired up
#   1. MongoDB container (via runMongoContainer.sh) - always run first
#   2. ships-go backend  (go run .)
#   3. ships-npc NPC controller (go run .)
#   4. ships-vue frontend (via runWeb.sh -> npm run dev)
#
# Shared config: `.env` and SSL certs are common to ships-go/ships-npc and
# ships-go/ships-vue respectively, so they live in one place, outside every
# project: ../files/.env and ../files/ssl/{cert.pem,key.pem} (sibling of
# ships-go/ships-npc/ships-vue). Each project's own ships-*/.env and
# ships-*/ssl are symlinks to those (gitignored, not committed - a fresh
# clone won't have them). This script:
#   - Creates ../files/ssl/{cert.pem,key.pem} (self-signed, localhost) if
#     missing.
#   - Recreates the ships-go/.env, ships-npc/.env, ships-go/ssl and
#     ships-vue/ssl symlinks if missing.
#   - Does NOT create ../files/.env: it holds secrets (MONGODB_URI,
#     NPC_SECRET, etc.) that must be provided manually.
#
# Logs: if run from inside a Konsole window, ships-go/ships-npc/ships-vue
# are each opened in their own new Konsole tab, so every project's log is
# visible on its own tab (Ctrl+C or closing a tab stops just that service).
# Otherwise (no Konsole, e.g. over SSH), they run in the background here and
# their output is both printed and saved to ./logs/*.log - open extra
# terminal tabs and `tail -f scripts/logs/<name>.log` to follow each one.
#
# Press Ctrl+C to stop ships-go, ships-npc and ships-vue started in THIS
# terminal (background mode). The MongoDB container is left running; use
# its own "exit" prompt (or `docker stop mongodb && docker rm mongodb`) to
# remove it.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

SHIPS_GO_DIR="$SCRIPT_DIR/.."
SHIPS_ROOT="$SCRIPT_DIR/../.."
FILES_DIR="$SHIPS_ROOT/files"

LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"

# ensure_symlink LINK_PATH RELATIVE_TARGET
# Creates LINK_PATH -> RELATIVE_TARGET if LINK_PATH doesn't exist yet.
# Leaves it untouched if it already exists (symlink or real file/dir).
ensure_symlink() {
    local link="$1" target="$2"
    if [ -e "$link" ] || [ -L "$link" ]; then
        return 0
    fi
    ln -s "$target" "$link"
    echo "  created $link -> $target"
}

ensure_shared_config() {
    mkdir -p "$FILES_DIR/ssl"

    if [ ! -f "$FILES_DIR/ssl/cert.pem" ] || [ ! -f "$FILES_DIR/ssl/key.pem" ]; then
        echo "No SSL certs found in $FILES_DIR/ssl, generating a self-signed dev certificate..."
        openssl req -x509 -newkey rsa:2048 -nodes -days 825 \
            -keyout "$FILES_DIR/ssl/key.pem" -out "$FILES_DIR/ssl/cert.pem" \
            -subj "/CN=localhost" \
            -addext "subjectAltName=DNS:localhost,IP:127.0.0.1" \
            2>/dev/null
        echo "  generated $FILES_DIR/ssl/cert.pem and key.pem"
    fi

    ensure_symlink "$SHIPS_GO_DIR/ssl" "../files/ssl"
    ensure_symlink "$SHIPS_ROOT/ships-vue/ssl" "../files/ssl"
    ensure_symlink "$SHIPS_GO_DIR/.env" "../files/.env"
    ensure_symlink "$SHIPS_ROOT/ships-npc/.env" "../files/.env"

    if [ ! -f "$FILES_DIR/.env" ]; then
        echo "WARNING: $FILES_DIR/.env does not exist yet. Create it with at least" \
            "MONGODB_URI, NPC_SECRET, SSL_CERT_PATH=./ssl/cert.pem and SSL_KEY_PATH=./ssl/key.pem" \
            "(see ships-go/README.md)."
    fi
}

PIDS=()
USE_TABS=false
if command -v konsole >/dev/null 2>&1 && [ -n "$KONSOLE_DBUS_SESSION" ]; then
    USE_TABS=true
fi

# run_service NAME DIR COMMAND...
# Opens a new Konsole tab running COMMAND (if available), otherwise runs it
# in the background, tee-ing its output to logs/NAME.log.
run_service() {
    local name="$1" dir="$2"
    shift 2
    if [ "$USE_TABS" = true ]; then
        konsole --new-tab -p tabtitle="$name" \
            -e bash -c "cd '$dir' && { $*; }; echo; echo '[$name] exited, press Enter to close this tab.'; read" &
        echo "  $name -> opened in a new Konsole tab"
    else
        (cd "$dir" && "$@") > >(tee "$LOG_DIR/$name.log") 2>&1 &
        PIDS+=($!)
        echo "  $name -> background, logging to $LOG_DIR/$name.log"
    fi
}

cleanup() {
    echo ""
    echo "Stopping services started in this terminal..."
    for pid in "${PIDS[@]}"; do
        kill "$pid" 2>/dev/null
    done
    wait "${PIDS[@]}" 2>/dev/null
    echo "Dev environment stopped (MongoDB container left running)."
}
trap cleanup EXIT INT TERM

echo "== Ensuring shared config (files/.env, files/ssl) =="
ensure_shared_config

echo "== Starting MongoDB via runMongoContainer.sh =="
./runMongoContainer.sh

echo "Waiting for MongoDB to accept connections..."
until docker exec mongodb mongosh \
    -u admin -p admin --authenticationDatabase admin \
    --eval "db.runCommand({ ping: 1 })" &>/dev/null; do
    sleep 1
done
echo "MongoDB is up."

echo "== Starting ships-go, ships-npc and ships-vue =="
run_service "ships-go" ".." go run .
run_service "ships-npc" "../../ships-npc" go run .
run_service "ships-vue" "." bash ./runWeb.sh

echo ""
if [ "$USE_TABS" = true ]; then
    echo "Dev environment is up: check the new Konsole tabs for each project's log."
    echo "This tab is now free; MongoDB was already started above."
else
    echo "Dev environment is up: ships-go, ships-npc and ships-vue are running in the"
    echo "background, logging to $LOG_DIR/*.log. Open new terminal tabs and run e.g.:"
    echo "  tail -f $LOG_DIR/ships-go.log"
    echo "Press Ctrl+C here to stop everything."
    wait "${PIDS[@]}"
fi
