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
# Logs: ships-go/ships-npc/ships-vue run in the background here, their
# output is saved to ./logs/*.log, and this script opens viewLogs.sh: a
# tabbed console viewer showing one project's log at a time with a tab bar,
# pinned to the top line, naming the one you are looking at, so the three
# streams are no longer interleaved into an unreadable mess. It works the
# same in every terminal - a plain console, over SSH or VS Code's
# integrated terminal.
#   [1-3] pick a project   [Tab] next   [a] merged view
#   [PgUp/PgDn] scroll back without losing the tab bar   [End] live again
#   [d]   detach (services keep running)   [q] quit and stop them
# Run scripts/viewLogs.sh on its own at any time to (re)attach.
# Pass --no-ui to get the plain interleaved output instead, or
# --konsole-tabs (inside Konsole only) to open one Konsole tab per project.
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
DETACHED=false
USE_TABS=false
USE_LOG_UI=true

# --no-ui restores the old behaviour (plain interleaved output), which is
# what you want when piping this script's output somewhere or running it
# from a non-interactive context like CI.
for arg in "$@"; do
    case "$arg" in
        --no-ui) USE_LOG_UI=false ;;
        --konsole-tabs) USE_TABS=true ;;
    esac
done

if [ "$USE_TABS" = true ] &&
    ! { command -v konsole >/dev/null 2>&1 && [ -n "${KONSOLE_DBUS_SESSION:-}" ]; }; then
    echo "  (--konsole-tabs ignored: not running inside Konsole)" >&2
    USE_TABS=false
fi
# The tabbed viewer is the default in every terminal, Konsole included, so
# the dev environment looks and works the same in a plain console, over SSH
# and in VS Code. It still needs a real terminal to draw on, and only makes
# sense when the services are logging here rather than into Konsole tabs.
if [ "$USE_TABS" = true ] || [ ! -t 0 ] || [ ! -t 1 ]; then
    USE_LOG_UI=false
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
    elif [ "$USE_LOG_UI" = true ]; then
        # The viewer owns the terminal, so the service must write to its log
        # only - tee-ing to stdout as well would scribble over the UI.
        #
        # setsid puts the service in its own process group so cleanup can
        # signal the whole tree. `go run .` compiles to a temporary binary
        # and execs it as a child, so killing just the pid recorded here
        # leaves that binary running - which is how a stray ships-npc once
        # survived a restart and ended up fighting the new one for control
        # of the NPCs.
        setsid bash -c "cd '$dir' && exec \"\$@\"" _ "$@" > "$LOG_DIR/$name.log" 2>&1 &
        PIDS+=($!)
        echo "  $name -> logging to $LOG_DIR/$name.log"
    else
        setsid bash -c "cd '$dir' && exec \"\$@\"" _ "$@" > >(tee "$LOG_DIR/$name.log") 2>&1 &
        PIDS+=($!)
        echo "  $name -> background, logging to $LOG_DIR/$name.log"
    fi
}

cleanup() {
    if [ "$DETACHED" = true ]; then
        return 0
    fi
    if [ ${#PIDS[@]} -eq 0 ]; then
        # Konsole mode: each service owns its tab and nothing was started in
        # the background here. A bare `wait` would not be a no-op - it waits
        # for *every* child, i.e. the konsole tabs, and never returns.
        return 0
    fi
    echo ""
    echo "Stopping services started in this terminal..."
    for pid in "${PIDS[@]}"; do
        # Negative pid = the whole process group (see setsid in run_service),
        # so `go run`'s compiled child dies with it instead of being orphaned.
        kill -- "-$pid" 2>/dev/null || kill "$pid" 2>/dev/null || true
    done
    # `wait` reports the signal that killed each service (143 = SIGTERM),
    # which under `set -e` would abort this handler before it finishes.
    wait "${PIDS[@]}" 2>/dev/null || true
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
elif [ "$USE_LOG_UI" = true ]; then
    echo "Dev environment is up. Opening the tabbed log viewer..."
    echo "  [1-3] pick a project   [a] all   [PgUp/PgDn] scroll   [d] detach   [q] quit + stop"
    sleep 1
    # Foreground, so this terminal becomes the viewer. It exits 10 when the
    # user detaches, which must leave the services running.
    set +e
    ./viewLogs.sh --log-dir "$LOG_DIR" ships-go ships-npc ships-vue
    viewer_status=$?
    set -e
    if [ "$viewer_status" -eq 10 ]; then
        DETACHED=true
        echo "Detached. ships-go, ships-npc and ships-vue are still running."
        echo "  reattach: $SCRIPT_DIR/viewLogs.sh"
        echo "  stop:     for p in ${PIDS[*]}; do kill -- -\$p; done"
        exit 0
    fi
else
    echo "Dev environment is up: ships-go, ships-npc and ships-vue are running in the"
    echo "background, logging to $LOG_DIR/*.log. Open new terminal tabs and run e.g.:"
    echo "  tail -f $LOG_DIR/ships-go.log"
    echo "  (or $SCRIPT_DIR/viewLogs.sh for a tabbed view of all three)"
    echo "Press Ctrl+C here to stop everything."
    wait "${PIDS[@]}" || true
fi
