#!/bin/bash
# Tabbed log viewer for the Ships dev environment.
#
# runDevEnvironment.sh saves every service's output to scripts/logs/*.log,
# but all three streams also land in the one terminal at once, which is
# unreadable - a Vite rebuild, a Go panic and ships-npc's reconnect chatter
# interleaved line by line. This shows one project at a time, with a tab bar
# naming the log you are looking at.
#
# It is deliberately dependency-free (no tmux/screen): plain bash + ANSI, so
# it works the same in VS Code's integrated terminal, over SSH or in any
# xterm-compatible console.
#
# The tab bar is pinned to the top line and must never be lost, so:
#   - row 1 is kept outside the scrolling region (DECSTBM), which stops the
#     tail output from ever pushing it off screen;
#   - it is repainted on a timer, so a stray escape sequence in a log line
#     cannot leave a stale or blank bar behind;
#   - the viewer runs on the alternate screen, so the terminal's own
#     scrollback can't scroll the bar out of view (and the shell's content
#     comes back untouched on exit). Scrolling back through a log is done
#     with the keys below instead, which keeps the bar on screen.
#
# Usage:
#   ./viewLogs.sh                  # ships-go / ships-npc / ships-vue
#   ./viewLogs.sh ships-go mongo   # only these logs, in this order
#   ./viewLogs.sh --log-dir DIR    # read logs from somewhere else
#
# Keys:
#   1..9        jump to a tab            Tab / -> / n   next tab
#   a           merged view (all logs)   Shift+Tab / <- / p   previous tab
#   Up/Down     scroll a line            PgUp/PgDn      scroll a page
#   Home        oldest line              End / f / G    back to live tail
#   c           clear the view           r   reload the current log
#   d           detach (leave services running)
#   q           quit
#
# Exit codes: 0 = quit, 10 = detach. runDevEnvironment.sh uses the
# difference to decide whether to stop the services it started.

set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
SERVICES=()

while [ $# -gt 0 ]; do
    case "$1" in
        --log-dir)
            LOG_DIR="$2"
            shift 2
            ;;
        -h|--help)
            sed -n '2,44p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
            exit 0
            ;;
        *)
            SERVICES+=("$1")
            shift
            ;;
    esac
done

if [ ${#SERVICES[@]} -eq 0 ]; then
    SERVICES=("ships-go" "ships-npc" "ships-vue")
fi

# One colour per service, reused by the tab bar and by the merged view's
# line prefixes so a tab and its lines are recognisably the same thing.
COLORS=("36" "33" "35" "32" "34" "31")

TAB_COUNT=${#SERVICES[@]}
ALL_TAB=$TAB_COUNT          # the merged view sits after the real logs
current=0
tail_pids=()
DETACH_EXIT=10
exit_code=0

# Terminal size is cached and only recomputed on SIGWINCH: the header is
# repainted twice a second and forking tput each time would be wasteful.
COLS=80
ROWS=24
read_size() {
    # stty asks the terminal driver directly. tput would rather trust a
    # LINES/COLUMNS pair inherited from the environment, which goes stale the
    # moment the window is resized and leaves the tab bar drawn for a size
    # the terminal no longer has.
    local size rows cols
    size="$(stty size 2>/dev/null || true)"
    rows="${size%% *}"
    cols="${size##* }"
    [[ "$rows" =~ ^[0-9]+$ ]] || rows="$(tput lines 2>/dev/null || echo 24)"
    [[ "$cols" =~ ^[0-9]+$ ]] || cols="$(tput cols 2>/dev/null || echo 80)"
    ROWS="$rows"
    COLS="$cols"
    [ "$ROWS" -gt 2 ] || ROWS=3
    [ "$COLS" -gt 20 ] || COLS=20
}

# follow=true means a tail is streaming into the view. Scrolling back stops
# the tail and pages through the file instead, so the view stays still while
# the service keeps logging; scroll_offset counts lines up from the newest.
follow=true
scroll_offset=0

# The merged view has no single file to page through, so its lines are also
# captured here as they are rendered.
BUF_DIR="$(mktemp -d "${TMPDIR:-/tmp}/viewLogs.XXXXXX")"
MERGED_BUF="$BUF_DIR/merged.log"

log_file() { printf '%s/%s.log' "$LOG_DIR" "${SERVICES[$1]}"; }
color_of() { printf '%s' "${COLORS[$(($1 % ${#COLORS[@]}))]}"; }

# The file the current tab pages through when scrolled back.
source_file() {
    if [ "$current" -eq "$ALL_TAB" ]; then
        printf '%s' "$MERGED_BUF"
    else
        log_file "$current"
    fi
}

stop_tails() {
    local pid
    for pid in "${tail_pids[@]:-}"; do
        [ -n "$pid" ] || continue
        # The tails are pipelines, so kill the whole process group entry:
        # killing only the head would leave a stray sed behind.
        kill "$pid" 2>/dev/null
        wait "$pid" 2>/dev/null
    done
    tail_pids=()
}

# draw_header repaints the tab bar on the top line. The cursor is saved and
# restored so a tail writing into the scrolling region below is unaffected,
# and the scrolling region is re-asserted every time: a log line carrying a
# stray reset would otherwise let the next line overwrite the bar.
draw_header() {
    local bar="" width=0 i name label status

    for ((i = 0; i < TAB_COUNT; i++)); do
        name="${SERVICES[$i]}"
        label=" $((i + 1)):$name "
        if [ "$i" -eq "$current" ]; then
            bar+=$'\e[7;1;'"$(color_of "$i")"'m'"$label"$'\e[0m'
        else
            bar+=$'\e['"$(color_of "$i")"'m'"$label"$'\e[0m'
        fi
        width=$((width + ${#label}))
    done
    label=" a:ALL "
    if [ "$current" -eq "$ALL_TAB" ]; then
        bar+=$'\e[7;1m'"$label"$'\e[0m'
    else
        bar+="$label"
    fi
    width=$((width + ${#label}))

    # The right-hand side says whether you are looking at live output or at a
    # frozen window of the file, which is the thing that is easy to lose
    # track of once you have scrolled back.
    if [ "$follow" = true ]; then
        status=" LIVE  [Tab] switch  [PgUp] scroll  [d]etach  [q]uit "
    else
        status=" SCROLLED -${scroll_offset}  [End] live  [d]etach  [q]uit "
    fi

    printf '\e7'                 # save cursor
    printf '\e[1;1H\e[2K'        # top line, cleared
    printf '%s' "$bar"
    if [ $((width + ${#status})) -lt "$COLS" ]; then
        if [ "$follow" = true ]; then
            printf '\e[1;%dH\e[2m%s\e[0m' $((COLS - ${#status} + 1)) "$status"
        else
            printf '\e[1;%dH\e[1;33m%s\e[0m' $((COLS - ${#status} + 1)) "$status"
        fi
    fi
    printf '\e[2;%dr' "$ROWS"    # keep row 1 out of the scrolling region
    printf '\e8'                 # restore cursor
}

# start_tail follows whichever log the current tab points at. The merged
# view follows all of them at once, prefixing each line with its service so
# the interleaving stays readable - which is the one case where mixing the
# streams is actually wanted.
start_tail() {
    local i name file
    if [ "$current" -eq "$ALL_TAB" ]; then
        for ((i = 0; i < TAB_COUNT; i++)); do
            name="${SERVICES[$i]}"
            file="$(log_file "$i")"
            [ -f "$file" ] || continue
            # Redirect into the prefixer instead of piping, so $! is the
            # tail's own pid: with a pipeline it would be sed's, and killing
            # sed leaves the tail alive until its next write (which never
            # comes for an idle log). The prefixed lines are also kept in
            # MERGED_BUF so this view can be scrolled back like the others.
            tail -n 20 -F "$file" 2>/dev/null \
                > >(sed -u "s/^/$(printf '\e[%sm%-10s\e[0m| ' "$(color_of "$i")" "$name")/" \
                    | tee -a "$MERGED_BUF") &
            tail_pids+=($!)
        done
        return
    fi

    file="$(log_file "$current")"
    if [ ! -f "$file" ]; then
        printf '\e[2m(waiting for %s ...)\e[0m\n' "$file"
    fi
    # -F (not -f) keeps waiting for a log that does not exist yet and follows
    # it across truncation, so attaching before a service has produced any
    # output - or after a restart rewrote the file - still works.
    tail -n 200 -F "$file" 2>/dev/null &
    tail_pids+=($!)
}

show_tab() {
    stop_tails
    follow=true
    scroll_offset=0
    if [ "$current" -eq "$ALL_TAB" ]; then
        : > "$MERGED_BUF"        # the merged view is rebuilt from scratch
    fi
    printf '\e[2J\e[2;1H'        # clear, cursor to the top of the region
    draw_header
    printf '\e[2;1H'
    start_tail
}

# render_page paints a fixed window of the current log instead of following
# it. Scrolling the terminal itself would take the tab bar with it (and the
# alternate screen has no scrollback anyway), so the viewer does the
# scrolling and the bar simply stays where it is.
render_page() {
    local src view total top text
    src="$(source_file)"
    view=$((ROWS - 1))
    total=0
    [ -f "$src" ] && total=$(wc -l < "$src" 2>/dev/null || echo 0)

    local max_offset=$((total - view))
    [ "$max_offset" -lt 0 ] && max_offset=0
    [ "$scroll_offset" -gt "$max_offset" ] && scroll_offset=$max_offset
    [ "$scroll_offset" -lt 0 ] && scroll_offset=0

    top=$((total - view + 1 - scroll_offset))
    [ "$top" -lt 1 ] && top=1

    text=""
    if [ -f "$src" ] && [ "$total" -gt 0 ]; then
        # Long lines are trimmed so one log line always occupies one row and
        # the window cannot drift; lines carrying escape sequences are left
        # alone rather than risk cutting one in half.
        text="$(sed -n "${top},$((top + view - 1))p" "$src" |
            awk -v w="$COLS" '{ if (index($0, "\033") == 0 && length($0) > w) print substr($0, 1, w); else print }')"
    fi

    printf '\e[2J\e[2;1H'
    # No trailing newline: printing one on the last row would scroll the
    # region by a line and desynchronise the window.
    printf '%s' "$text"
    draw_header
}

# scroll_by moves the window; the first movement away from the bottom stops
# the tail so the text underneath stops moving while you read it.
scroll_by() {
    local delta="$1"
    if [ "$follow" = true ]; then
        [ "$delta" -le 0 ] && return   # already at the newest line
        stop_tails
        follow=false
        scroll_offset=0
    fi
    scroll_offset=$((scroll_offset + delta))
    [ "$scroll_offset" -lt 0 ] && scroll_offset=0
    if [ "$scroll_offset" -eq 0 ]; then
        resume_follow
        return
    fi
    render_page
}

scroll_to_top() {
    if [ "$follow" = true ]; then
        stop_tails
        follow=false
    fi
    scroll_offset=$(wc -l < "$(source_file)" 2>/dev/null || echo 0)
    render_page
}

resume_follow() {
    stop_tails
    follow=true
    scroll_offset=0
    printf '\e[2J\e[2;1H'
    draw_header
    printf '\e[2;1H'
    start_tail
}

select_tab() {
    local wanted="$1"
    [ "$wanted" -eq "$current" ] && return
    current="$wanted"
    show_tab
}

next_tab() { select_tab $(((current + 1) % (TAB_COUNT + 1))); }
prev_tab() { select_tab $(((current - 1 + TAB_COUNT + 1) % (TAB_COUNT + 1))); }

on_resize() {
    read_size
    if [ "$follow" = true ]; then
        draw_header
    else
        render_page
    fi
}

restore_terminal() {
    stop_tails
    printf '\e[r'                # drop the scrolling region
    printf '\e[?25h'             # cursor back on
    printf '\e[?1049l'           # back to the normal screen, shell intact
    stty "$saved_stty" 2>/dev/null
    rm -rf "$BUF_DIR"
}

if [ ! -d "$LOG_DIR" ]; then
    echo "No log directory at $LOG_DIR" >&2
    exit 1
fi

saved_stty="$(stty -g 2>/dev/null)"
trap restore_terminal EXIT
trap on_resize WINCH

# `read -s` only suppresses echo while it is actually reading; a key pressed
# while the viewer is repainting would otherwise be echoed into the log view
# as literal noise (^[[A for an arrow key, for instance).
stty -echo 2>/dev/null

read_size
printf '\e[?1049h'               # alternate screen: nothing to scroll away
show_tab

while true; do
    # A short timeout keeps the loop responsive to WINCH and lets the tails
    # keep writing while we wait for a keypress. Every timeout also repaints
    # the header, so the tab bar is restored within half a second even if a
    # log line managed to scribble over it.
    if ! IFS= read -rsn1 -t 0.5 key; then
        draw_header
        continue
    fi

    case "$key" in
        $'\e')
            # Arrow/PgUp/Home keys arrive as escape sequences of different
            # lengths, so read until the final byte instead of assuming two
            # characters; a bare Esc simply times out and is ignored.
            seq=""
            while [ ${#seq} -lt 8 ]; do
                IFS= read -rsn1 -t 0.05 ch || break
                seq+="$ch"
                case "$ch" in [A-Za-z~]) break ;; esac
            done
            case "$seq" in
                '[C') next_tab ;;                    # right
                '[D') prev_tab ;;                    # left
                '[Z') prev_tab ;;                    # shift+tab
                '[A') scroll_by 1 ;;                 # up
                '[B') scroll_by -1 ;;                # down
                '[5~') scroll_by $((ROWS - 2)) ;;    # page up
                '[6~') scroll_by $((2 - ROWS)) ;;    # page down
                '[H'|'[1~') scroll_to_top ;;         # home
                '[F'|'[4~') resume_follow ;;         # end
                *) : ;;
            esac
            ;;
        $'\t') next_tab ;;
        [1-9])
            idx=$((key - 1))
            [ "$idx" -lt "$TAB_COUNT" ] && select_tab "$idx"
            ;;
        a|A) select_tab "$ALL_TAB" ;;
        n|N) next_tab ;;
        p|P) prev_tab ;;
        f|F|G) resume_follow ;;
        g) scroll_to_top ;;
        c|C)
            printf '\e[2J\e[2;1H'
            draw_header
            printf '\e[2;1H'
            ;;
        r|R) show_tab ;;
        d|D)
            exit_code=$DETACH_EXIT
            break
            ;;
        q|Q) break ;;
        *) : ;;
    esac
done

exit "$exit_code"
